package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
)

// fakeDoer records calls and returns canned envelope data, letting tests assert
// the exact multipart fields without any HTTP.
type fakeDoer struct {
	lastBaseURL string
	lastPath    string
	lastFields  map[string]string
	respData    string // JSON unmarshaled into out
	rawResp     []byte // returned by DoMultipartRaw
	err         error
}

func (f *fakeDoer) DoMultipart(_ context.Context, baseURL, path string, fields map[string]string, out any) error {
	f.lastBaseURL = baseURL
	f.lastPath = path
	f.lastFields = fields
	if f.err != nil {
		return f.err
	}
	if out != nil && f.respData != "" {
		return json.Unmarshal([]byte(f.respData), out)
	}
	return nil
}

func (f *fakeDoer) DoMultipartRaw(_ context.Context, baseURL, path string, fields map[string]string) ([]byte, error) {
	f.lastBaseURL = baseURL
	f.lastPath = path
	f.lastFields = fields
	if f.err != nil {
		return nil, f.err
	}
	return f.rawResp, nil
}

func (f *fakeDoer) DoFormURLEncoded(_ context.Context, _, _ string, _ url.Values, _ any) error {
	return nil
}

func (f *fakeDoer) DoFormURLEncodedRaw(_ context.Context, _, _ string, _ url.Values) ([]byte, error) {
	return nil, nil
}

func (f *fakeDoer) BaseURL() string         { return "http://base.test" }
func (f *fakeDoer) WebUnblockerURL() string { return "http://unblock.test" }
func (f *fakeDoer) ScraperURL() string      { return "http://scraper.test" }

func ctx() context.Context { return context.Background() }

func TestWhitelistAdd_Fields(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)

	err := svc.Whitelist.Add(ctx(), AddWhitelistParams{
		Product: ProductResidential,
		IP:      "10.10.10.1",
		Remark:  "test",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if d.lastBaseURL != "http://base.test" {
		t.Errorf("baseURL = %q, want general host", d.lastBaseURL)
	}
	if d.lastPath != "/v1/white_list/add" {
		t.Errorf("path = %q", d.lastPath)
	}
	want := map[string]string{"product": "1", "ip": "10.10.10.1", "remark": "test"}
	assertFields(t, d.lastFields, want)
}

func TestWhitelistAdd_OmitsEmptyRemark(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	if err := svc.Whitelist.Add(ctx(), AddWhitelistParams{Product: 1, IP: "1.2.3.4"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, ok := d.lastFields["remark"]; ok {
		t.Errorf("empty remark should be omitted, got %v", d.lastFields)
	}
}

func TestWhitelistList_Unwrap(t *testing.T) {
	d := &fakeDoer{respData: `{
		"list":[{"id":12,"uid":81,"mark_ip":"10.10.10.1","ip":168430081,
		         "status":1,"lock":false,"mark":"test"}],
		"total":1
	}`}
	svc := New(d)

	lock := 1
	res, err := svc.Whitelist.List(ctx(), ListWhitelistParams{
		Product: ProductStaticISP,
		IP:      "10.10.10.1",
		Lock:    &lock,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if d.lastFields["product"] != "5" || d.lastFields["lock"] != "1" {
		t.Errorf("fields = %v", d.lastFields)
	}
	if res.Total != 1 || len(res.List) != 1 {
		t.Fatalf("res = %+v", res)
	}
	if res.List[0].MarkIP != "10.10.10.1" || res.List[0].IP != 168430081 {
		t.Errorf("item = %+v", res.List[0])
	}
}

func TestWhitelistList_LockNilOmitted(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[],"total":0}`}
	svc := New(d)
	if _, err := svc.Whitelist.List(ctx(), ListWhitelistParams{Product: 1}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, ok := d.lastFields["lock"]; ok {
		t.Errorf("nil Lock should be omitted, got %v", d.lastFields)
	}
}

func TestWhitelistDelete_Fields(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	if err := svc.Whitelist.Delete(ctx(), DeleteWhitelistParams{
		Product: 1, IPs: "1.1.1.1,2.2.2.2",
	}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	assertFields(t, d.lastFields, map[string]string{"product": "1", "ips": "1.1.1.1,2.2.2.2"})
}

func TestAccountCreate_Fields(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	err := svc.Account.Create(ctx(), CreateAccountParams{
		Product:  ProductResidential,
		Account:  "account11",
		Password: "pass11",
		Status:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.lastPath != "/v1/proxy_account/create" {
		t.Errorf("path = %q", d.lastPath)
	}
	want := map[string]string{
		"product": "1", "account": "account11", "password": "pass11", "status": "1",
	}
	assertFields(t, d.lastFields, want)
}

func TestAccountList_DefaultsAndUnwrap(t *testing.T) {
	d := &fakeDoer{respData: `{
		"list":[{"id":2,"account":"account11","status":1,
		         "limit_residential_flow":1100000000}],
		"page":1,"total":2
	}`}
	svc := New(d)

	res, err := svc.Account.List(ctx(), ListAccountParams{Product: ProductResidential})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Page/Limit defaulted.
	if d.lastFields["page"] != "1" || d.lastFields["limit"] != "10" {
		t.Errorf("pagination defaults not applied: %v", d.lastFields)
	}
	if res.Total != 2 || len(res.List) != 1 {
		t.Fatalf("res = %+v", res)
	}
	if res.List[0].LimitResidentialFlow != 1100000000 {
		t.Errorf("flow = %d", res.List[0].LimitResidentialFlow)
	}
}

func TestAccountConsumeLog_Fields(t *testing.T) {
	d := &fakeDoer{respData: `{"list":[{"account_id":2,"day":1731340800}]}`}
	svc := New(d)
	res, err := svc.Account.ConsumeLog(ctx(), ConsumeLogParams{
		AccountID: 2,
		StartTime: "2025-01-01 00:00:00",
		Page:      3,
	})
	if err != nil {
		t.Fatalf("ConsumeLog: %v", err)
	}
	if d.lastFields["account_id"] != "2" || d.lastFields["page"] != "3" || d.lastFields["limit"] != "10" {
		t.Errorf("fields = %v", d.lastFields)
	}
	if d.lastFields["start_time"] != "2025-01-01 00:00:00" {
		t.Errorf("start_time = %q", d.lastFields["start_time"])
	}
	if len(res.List) != 1 || res.List[0].Day != 1731340800 {
		t.Fatalf("res = %+v", res)
	}
}

func TestAccountUpdate_Fields(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	err := svc.Account.Update(ctx(), UpdateAccountParams{
		ID:        2,
		Account:   "account11",
		Password:  "pass11",
		Status:    -3,
		Remark:    "disabled",
		LimitFlow: "5",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if d.lastPath != "/v1/proxy_account/update" {
		t.Errorf("path = %q", d.lastPath)
	}
	want := map[string]string{
		"id": "2", "account": "account11", "password": "pass11",
		"status": "-3", "remark": "disabled", "limit_flow": "5",
	}
	assertFields(t, d.lastFields, want)
}

func TestWhitelistRemark_Fields(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	if err := svc.Whitelist.Remark(ctx(), RemarkWhitelistParams{
		Product: ProductResidential, ID: "101", Remark: "some mark",
	}); err != nil {
		t.Fatalf("Remark: %v", err)
	}
	assertFields(t, d.lastFields, map[string]string{
		"product": "1", "id": "101", "remark": "some mark",
	})
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Method: "Whitelist.Add", Fields: []string{"product", "ip"}}
	got := err.Error()
	if got != "proxy: Whitelist.Add: missing required field(s): product, ip" {
		t.Errorf("Error() = %q", got)
	}
}

func TestValidation_MissingRequired(t *testing.T) {
	svc := New(&fakeDoer{})
	tests := []struct {
		name   string
		call   func() error
		fields []string
	}{
		{"whitelist add", func() error {
			return svc.Whitelist.Add(ctx(), AddWhitelistParams{})
		}, []string{"product", "ip"}},
		{"account create", func() error {
			return svc.Account.Create(ctx(), CreateAccountParams{Account: "a"})
		}, []string{"product", "password", "status"}},
		{"account update", func() error {
			return svc.Account.Update(ctx(), UpdateAccountParams{})
		}, []string{"id", "account", "password"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("err = %v, want *ValidationError", err)
			}
			assertStringSet(t, ve.Fields, tc.fields)
		})
	}
}

func TestValidation_PreventsRequest(t *testing.T) {
	d := &fakeDoer{}
	svc := New(d)
	_ = svc.Whitelist.Add(ctx(), AddWhitelistParams{}) // invalid
	if d.lastPath != "" {
		t.Errorf("request should not have been sent, path = %q", d.lastPath)
	}
}

func TestList_PropagatesTransportError(t *testing.T) {
	d := &fakeDoer{err: errors.New("boom")}
	svc := New(d)
	if _, err := svc.Account.List(ctx(), ListAccountParams{Product: 1}); err == nil {
		t.Fatal("expected transport error to propagate")
	}
}

func assertFields(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("field count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("field %q = %q, want %q", k, got[k], v)
		}
	}
}

func assertStringSet(t *testing.T, got, want []string) {
	t.Helper()
	set := make(map[string]bool, len(got))
	for _, s := range got {
		set[s] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Errorf("missing field %q in %v", w, got)
		}
	}
	if len(got) != len(want) {
		t.Errorf("field set = %v, want %v", got, want)
	}
}
