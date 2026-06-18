package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
)

// fakeDoer records the last call and returns canned data.
type fakeDoer struct {
	mpPath   string
	mpFields map[string]string
	mpResp   string

	rawHost   string
	rawPath   string
	rawValues url.Values
	rawResp   []byte
	err       error

	feHost   string
	fePath   string
	feValues url.Values
	feResp   string
}

func (f *fakeDoer) DoMultipart(_ context.Context, _, path string, fields map[string]string, out any) error {
	f.mpPath = path
	f.mpFields = fields
	if f.err != nil {
		return f.err
	}
	if out != nil && f.mpResp != "" {
		return json.Unmarshal([]byte(f.mpResp), out)
	}
	return nil
}

func (f *fakeDoer) DoMultipartRaw(_ context.Context, _, path string, fields map[string]string) ([]byte, error) {
	f.mpPath = path
	f.mpFields = fields
	return f.rawResp, f.err
}

func (f *fakeDoer) DoFormURLEncoded(_ context.Context, host, path string, values url.Values, out any) error {
	f.feHost = host
	f.fePath = path
	f.feValues = values
	if f.err != nil {
		return f.err
	}
	if out != nil && f.feResp != "" {
		return json.Unmarshal([]byte(f.feResp), out)
	}
	return nil
}

func (f *fakeDoer) DoFormURLEncodedRaw(_ context.Context, host, path string, values url.Values) ([]byte, error) {
	f.rawHost = host
	f.rawPath = path
	f.rawValues = values
	if f.err != nil {
		return nil, f.err
	}
	return f.rawResp, nil
}

func (f *fakeDoer) BaseURL() string         { return "http://base.test" }
func (f *fakeDoer) WebUnblockerURL() string { return "http://unblock.test" }
func (f *fakeDoer) ScraperURL() string      { return "http://scraper.test" }

func ctx() context.Context { return context.Background() }

func TestDo_EncodesAndRoutesToScraperAPI(t *testing.T) {
	d := &fakeDoer{rawResp: []byte(`{"ok":true}`)}
	res, err := New(d).Do(ctx(), Request{
		Target:       TargetScraperAPI,
		ScraperName:  "youtube.com",
		ScraperID:    "youtube_video-post_explore",
		Params:       []map[string]any{{"url": "https://youtu.be/x"}},
		ReturnErrors: true,
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if res.Raw != `{"ok":true}` {
		t.Errorf("raw = %q", res.Raw)
	}
	if d.rawHost != "http://scraper.test" || d.rawPath != "/request" {
		t.Errorf("routed to host=%q path=%q", d.rawHost, d.rawPath)
	}
	if d.rawValues.Get("scraper_name") != "youtube.com" ||
		d.rawValues.Get("scraper_id") != "youtube_video-post_explore" {
		t.Errorf("values = %v", d.rawValues)
	}
	if d.rawValues.Get("scraper_errors") != "true" {
		t.Errorf("scraper_errors = %q", d.rawValues.Get("scraper_errors"))
	}
	// scraper_params must be the JSON-marshaled array.
	var got []map[string]any
	if err := json.Unmarshal([]byte(d.rawValues.Get("scraper_params")), &got); err != nil {
		t.Fatalf("scraper_params not valid JSON: %v", err)
	}
	if len(got) != 1 || got[0]["url"] != "https://youtu.be/x" {
		t.Errorf("scraper_params = %q", d.rawValues.Get("scraper_params"))
	}
}

func TestDo_RoutesToWebUnblocker(t *testing.T) {
	d := &fakeDoer{rawResp: []byte("ok")}
	_, err := New(d).Do(ctx(), Request{
		Target:      TargetWebUnblocker,
		ScraperName: "example.com",
		ScraperID:   "web-unlocker",
		Params:      []map[string]any{{"url": "https://example.com"}},
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if d.rawHost != "http://unblock.test" {
		t.Errorf("host = %q, want web unblocker", d.rawHost)
	}
	if d.rawValues.Get("scraper_errors") != "" {
		t.Errorf("scraper_errors should be unset, got %q", d.rawValues.Get("scraper_errors"))
	}
}

func TestDo_Validation(t *testing.T) {
	_, err := New(&fakeDoer{}).Do(ctx(), Request{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
	want := map[string]bool{"scraper_name": true, "scraper_id": true, "scraper_params": true}
	if len(ve.Fields) != 3 {
		t.Errorf("fields = %v", ve.Fields)
	}
	for _, f := range ve.Fields {
		if !want[f] {
			t.Errorf("unexpected missing field %q", f)
		}
	}
}

func TestDo_PropagatesError(t *testing.T) {
	d := &fakeDoer{err: errors.New("boom")}
	_, err := New(d).Do(ctx(), Request{
		ScraperName: "a", ScraperID: "b", Params: []map[string]any{{"x": 1}},
	})
	if err == nil {
		t.Fatal("expected error to propagate")
	}
}

func TestYouTubeVideoPost(t *testing.T) {
	d := &fakeDoer{rawResp: []byte(`{"title":"t"}`)}
	res, err := New(d).API.YouTube.VideoPost(ctx(), YouTubeVideoParams{
		URL: "https://www.youtube.com/watch?v=HAwTwmzgNc4",
	})
	if err != nil {
		t.Fatalf("VideoPost: %v", err)
	}
	if res.Raw != `{"title":"t"}` {
		t.Errorf("raw = %q", res.Raw)
	}
	if d.rawHost != "http://scraper.test" || d.rawPath != "/request" {
		t.Errorf("host=%q path=%q", d.rawHost, d.rawPath)
	}
	if d.rawValues.Get("scraper_id") != "youtube_video-post_explore" {
		t.Errorf("scraper_id = %q", d.rawValues.Get("scraper_id"))
	}
	var p []map[string]any
	_ = json.Unmarshal([]byte(d.rawValues.Get("scraper_params")), &p)
	if p[0]["url"] != "https://www.youtube.com/watch?v=HAwTwmzgNc4" {
		t.Errorf("params = %v", p)
	}
}

func TestUniversal_Balance(t *testing.T) {
	d := &fakeDoer{mpResp: `{"scraper_balance":17056}`}
	res, err := New(d).Universal.Balance(ctx())
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if d.mpPath != "/v1/capture/get_balance" || res.ScraperBalance != 17056 {
		t.Errorf("path=%q res=%+v", d.mpPath, res)
	}
}

func TestUniversal_Logs(t *testing.T) {
	d := &fakeDoer{mpResp: `{"list":[{"time_label":"2026-03-26","scraper_total_cost":1.5}]}`}
	res, err := New(d).Universal.Logs(ctx())
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if d.mpPath != "/v1/capture/logs" || len(res.List) != 1 || res.List[0].ScraperTotalCost != 1.5 {
		t.Errorf("path=%q res=%+v", d.mpPath, res)
	}
}

func TestUniversal_Unit(t *testing.T) {
	d := &fakeDoer{mpResp: `{"scraper":[{"package":"scraper","level":1,"price":1.6}],"unblocker":[{"package":"web_unlocker","level":1,"price":1.3}]}`}
	res, err := New(d).Universal.Unit(ctx())
	if err != nil {
		t.Fatalf("Unit: %v", err)
	}
	if d.mpPath != "/v1/capture/unit" || res.Scraper[0].Price != 1.6 || res.Unblocker[0].Price != 1.3 {
		t.Errorf("res = %+v", res)
	}
}

func TestUnblocker_Areas(t *testing.T) {
	d := &fakeDoer{mpResp: `{"continent":{"1":"Asia"},"country":[{"continent":"Asia","continent_code":1,"list":[{"code":"jp","name_en":"Japan"}]}]}`}
	res, err := New(d).Unblocker.Countries(ctx())
	if err != nil {
		t.Fatalf("Countries: %v", err)
	}
	if d.mpPath != "/v1/proxy/unblocker_area" || res.Country[0].List[0].Code != "jp" {
		t.Errorf("path=%q res=%+v", d.mpPath, res)
	}
}

func TestUnblocker_StatesCitiesISPs(t *testing.T) {
	ds := &fakeDoer{mpResp: `{"data":[{"state":"hialeah"}]}`}
	st, err := New(ds).Unblocker.States(ctx(), "us")
	if err != nil || ds.mpPath != "/v1/proxy/unblocker_area_by_country" || st.Data[0].State != "hialeah" {
		t.Fatalf("States: %v %+v", err, st)
	}

	dc := &fakeDoer{mpResp: `{"data":[{"code":"guntersville"}]}`}
	ci, err := New(dc).Unblocker.Cities(ctx(), "us", "Firstmesa")
	if err != nil || dc.mpFields["region"] != "Firstmesa" || ci.Data[0].Code != "guntersville" {
		t.Fatalf("Cities: %v %+v", err, ci)
	}

	di := &fakeDoer{mpResp: `{"list":[{"asn":"AS1","show_name":"X"}]}`}
	isp, err := New(di).Unblocker.ISPs(ctx(), "SG")
	if err != nil || di.mpPath != "/v1/proxy/unblocker_city_isp" || isp.List[0].ASN != "AS1" {
		t.Fatalf("ISPs: %v %+v", err, isp)
	}
}

func TestUnblocker_Validation(t *testing.T) {
	if _, err := New(&fakeDoer{}).Unblocker.States(ctx(), ""); err == nil {
		t.Error("States empty code should fail")
	}
	_, err := New(&fakeDoer{}).Unblocker.Cities(ctx(), "", "")
	var ve *ValidationError
	if !errors.As(err, &ve) || len(ve.Fields) != 2 {
		t.Errorf("Cities validation = %v", err)
	}
}

func TestUnblocker_Scrape(t *testing.T) {
	d := &fakeDoer{feResp: `{"code":200,"html":"<html></html>","msg_detail":"","use_balance":0.0013}`}
	js := true
	res, err := New(d).Unblocker.Scrape(ctx(), UnblockerParams{
		TargetURL: "https://www.google.com",
		Country:   "us",
		JSRender:  &js,
		WaitMS:    5000,
	})
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if d.feHost != "http://unblock.test" || d.fePath != "/request" {
		t.Errorf("routed to host=%q path=%q", d.feHost, d.fePath)
	}
	if d.feValues.Get("target_url") != "https://www.google.com" ||
		d.feValues.Get("response_format") != "html" || // defaulted
		d.feValues.Get("js_render") != "true" ||
		d.feValues.Get("country") != "us" ||
		d.feValues.Get("wait_ms") != "5000" {
		t.Errorf("values = %v", d.feValues)
	}
	if res.Code != 200 || res.HTML != "<html></html>" || res.UseBalance != 0.0013 {
		t.Errorf("result = %+v", res)
	}
}

func TestUnblocker_ScrapeValidation(t *testing.T) {
	_, err := New(&fakeDoer{}).Unblocker.Scrape(ctx(), UnblockerParams{})
	var ve *ValidationError
	if !errors.As(err, &ve) || len(ve.Fields) != 1 || ve.Fields[0] != "target_url" {
		t.Errorf("validation = %v", err)
	}
}

func TestBrowser_CountriesAndFlow(t *testing.T) {
	dc := &fakeDoer{mpResp: `{"x":1}`}
	raw, err := New(dc).Browser.Countries(ctx())
	if err != nil || dc.mpPath != "/v1/proxy/browser_area" || string(raw) != `{"x":1}` {
		t.Fatalf("Countries: %v raw=%s", err, raw)
	}

	df := &fakeDoer{mpResp: `{"list":[{"id":1,"use":92295}]}`}
	fl, err := New(df).Browser.FlowUse(ctx(), "2025-01-01 00:00:00", "2025-01-02 00:00:00")
	if err != nil {
		t.Fatalf("FlowUse: %v", err)
	}
	if df.mpPath != "/v1/browser_flow/browser_flow_use" || fl.List[0].Use != 92295 {
		t.Errorf("path=%q res=%+v", df.mpPath, fl)
	}

	if _, err := New(&fakeDoer{}).Browser.FlowUse(ctx(), "", ""); err == nil {
		t.Error("FlowUse empty bounds should fail")
	}
}

func TestValidationError_Message(t *testing.T) {
	e := &ValidationError{Fields: []string{"scraper_name", "scraper_id"}}
	if e.Error() != "scraper: missing required field(s): scraper_name, scraper_id" {
		t.Errorf("Error() = %q", e.Error())
	}
}
