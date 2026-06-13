package wallet

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
)

type fakeDoer struct {
	path   string
	fields map[string]string
	resp   string
}

func (f *fakeDoer) DoMultipart(_ context.Context, _, path string, fields map[string]string, out any) error {
	f.path = path
	f.fields = fields
	if out != nil && f.resp != "" {
		return json.Unmarshal([]byte(f.resp), out)
	}
	return nil
}

func (f *fakeDoer) DoMultipartRaw(_ context.Context, _, _ string, _ map[string]string) ([]byte, error) {
	return nil, nil
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

func TestBalance(t *testing.T) {
	d := &fakeDoer{resp: `{"balance":104}`}
	res, err := New(d).Balance(context.Background())
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if d.path != "/v1/wallet/balance" || res.Balance != 104 {
		t.Errorf("path=%q res=%+v", d.path, res)
	}
}

func TestUsageRecord_DefaultsAndUnwrap(t *testing.T) {
	d := &fakeDoer{resp: `{"count":37,"list":[{"id":845,"order_type":"static_open","money":0,"pay_type":"Wallet","description":"StaticIp"}]}`}
	res, err := New(d).UsageRecord(context.Background(), UsageRecordParams{})
	if err != nil {
		t.Fatalf("UsageRecord: %v", err)
	}
	if d.path != "/v1/wallet/usage_record" {
		t.Errorf("path = %q", d.path)
	}
	if d.fields["page"] != "1" || d.fields["limit"] != "10" {
		t.Errorf("defaults not applied: %v", d.fields)
	}
	if res.Count != 37 || res.List[0].ID != 845 || res.List[0].PayType != "Wallet" {
		t.Errorf("res = %+v", res)
	}
}

func TestUsageRecord_CustomPaging(t *testing.T) {
	d := &fakeDoer{resp: `{"count":0,"list":[]}`}
	if _, err := New(d).UsageRecord(context.Background(), UsageRecordParams{Page: 3, Limit: 50}); err != nil {
		t.Fatalf("UsageRecord: %v", err)
	}
	if d.fields["page"] != "3" || d.fields["limit"] != "50" {
		t.Errorf("paging = %v", d.fields)
	}
}
