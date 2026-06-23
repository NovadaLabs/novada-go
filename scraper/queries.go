package scraper

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// --- Universal -------------------------------------------------------------

// UniversalService exposes the shared scraper account queries (balance, usage
// logs and unit prices) on the general host.
type UniversalService struct {
	d transport.Doer
}

// ScraperBalance is the remaining scraper balance.
type ScraperBalance struct {
	ScraperBalance float64 `json:"scraper_balance"`
}

// UsageLog is a single daily scraper/unblocker/browser consumption record.
type UsageLog struct {
	TimeLabel         string  `json:"time_label"`
	UnlockerTotalCost float64 `json:"unlocker_total_cost"`
	ScraperTotalCost  float64 `json:"scraper_total_cost"`
	UnlockerUsedRes   int64   `json:"unlocker_used_res"`
	ScraperUsedRes    int64   `json:"scraper_used_res"`
	ScraperUsedFlow   int64   `json:"scraper_used_flow"`
	BrowserTotalCost  float64 `json:"browser_total_cost"`
	BrowserUsedFlow   int64   `json:"browser_used_flow"`
}

// UsageLogList is the data payload of Logs.
type UsageLogList struct {
	List []UsageLog `json:"list"`
}

// UnitPrice is a single price tier for a scraper package.
type UnitPrice struct {
	Package   string  `json:"package"`
	Level     int     `json:"level"`
	Price     float64 `json:"price"`
	Available int     `json:"available"`
}

// UnitPrices is the data payload of Unit, grouping prices by product.
type UnitPrices struct {
	Scraper   []UnitPrice `json:"scraper"`
	Unblocker []UnitPrice `json:"unblocker"`
}

// Balance returns the remaining scraper balance
// (POST /v1/capture/get_balance). The endpoint returns the balance as a bare
// number in the "data" field, so we decode into an int64 and wrap it.
func (s *UniversalService) Balance(ctx context.Context) (*ScraperBalance, error) {
	var bal float64
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/capture/get_balance", map[string]string{}, &bal); err != nil {
		return nil, err
	}
	return &ScraperBalance{ScraperBalance: bal}, nil
}

// Logs returns the scraper consumption records (POST /v1/capture/logs).
func (s *UniversalService) Logs(ctx context.Context) (*UsageLogList, error) {
	var out UsageLogList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/capture/logs", map[string]string{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Unit returns the user's capture unit prices (POST /v1/capture/unit).
func (s *UniversalService) Unit(ctx context.Context) (*UnitPrices, error) {
	var out UnitPrices
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/capture/unit", map[string]string{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Areas (shared shapes) -------------------------------------------------

// Areas is a country listing grouped by continent, shared by the unblocker
// area endpoint and the residential-style listings.
type Areas struct {
	Continent map[string]string `json:"continent"`
	Country   []ContinentGroup  `json:"country"`
}

// ContinentGroup is a continent with its countries.
type ContinentGroup struct {
	Continent     string    `json:"continent"`
	ContinentCode int       `json:"continent_code"`
	List          []Country `json:"list"`
}

// Country is a single country in an area listing.
type Country struct {
	Code      string `json:"code"`
	Continent int    `json:"continent"`
	IPNum     int    `json:"ip_num"`
	Name      string `json:"name"`
	NameEN    string `json:"name_en"`
	Protocol  int    `json:"protocol"`
}

// State is a state/region within a country.
type State struct {
	State string `json:"state"`
}

// StateList is the data payload of the state listings.
type StateList struct {
	Data []State `json:"data"`
}

// City is a city within a region.
type City struct {
	Code string `json:"code"`
}

// CityList is the data payload of the city listings.
type CityList struct {
	Data []City `json:"data"`
}

// ISP is a carrier available in a country.
type ISP struct {
	ASN      string `json:"asn"`
	ShowName string `json:"show_name"`
}

// ISPList is the data payload of the ISP listings.
type ISPList struct {
	List []ISP `json:"list"`
}

// --- Unblocker -------------------------------------------------------------

// UnblockerService exposes Web Unblocker scraping plus its area queries. The
// area queries are /v1/* APIs on the general host; the scrape itself uses
// Scrape (which routes to the Web Unblocker host).
type UnblockerService struct {
	d transport.Doer
}

// UnblockerParams configures a Web Unblocker scrape (POST /request). Only
// TargetURL is required; ResponseFormat defaults to "html" when empty, and the
// remaining optional fields are sent only when set.
type UnblockerParams struct {
	// TargetURL is the destination address to scrape. Required.
	TargetURL string
	// ResponseFormat is the output format: "html", "png" or "html,png"
	// (comma-separated). Defaults to "html" when empty.
	ResponseFormat string
	// JSRender enables JS rendering to capture dynamically loaded content.
	JSRender *bool
	// Headers are custom request headers used to access the site.
	Headers string
	// Cookies are custom cookies used to access the site.
	Cookies string
	// Country is the proxy country/region, e.g. "us".
	Country string
	// WaitMS is the maximum page wait time in milliseconds (max 100000). Sent
	// only when positive.
	WaitMS int
	// WaitSelector waits for a CSS selector to appear in the DOM (max 30s). It
	// takes precedence over WaitMS when both are set.
	WaitSelector string
	// FollowRedirects follows redirects from expired URLs to their new target.
	FollowRedirects *bool
	// BlockResources skips loading images, JS files and videos to speed up the
	// crawl.
	BlockResources *bool
	// Clear strips unnecessary JS and CSS from the crawl result.
	Clear *bool
	// AutoRuns is the number of automatic retries on proxy failure (0-10,
	// server default 2). Sent only when non-nil.
	AutoRuns *int
}

// UnblockerResult is the decoded data payload of a Web Unblocker scrape.
type UnblockerResult struct {
	// Code is the page-level status (200 on success).
	Code int `json:"code"`
	// HTML is the scraped content.
	HTML string `json:"html"`
	// Msg is a short status message.
	Msg string `json:"msg"`
	// MsgDetail carries additional error detail when the scrape fails.
	MsgDetail string `json:"msg_detail"`
	// UseBalance is the balance consumed by this call.
	UseBalance float64 `json:"use_balance"`
}

// Scrape submits a Web Unblocker scrape job (POST /request on the Web Unblocker
// host) and decodes the structured result. TargetURL is required.
func (s *UnblockerService) Scrape(ctx context.Context, p UnblockerParams) (*UnblockerResult, error) {
	if err := requireField("target_url", p.TargetURL); err != nil {
		return nil, err
	}

	format := strings.TrimSpace(p.ResponseFormat)
	if format == "" {
		format = "html"
	}

	values := url.Values{}
	values.Set("target_url", p.TargetURL)
	values.Set("response_format", format)
	if p.JSRender != nil {
		values.Set("js_render", strconv.FormatBool(*p.JSRender))
	}
	if p.Headers != "" {
		values.Set("headers", p.Headers)
	}
	if p.Cookies != "" {
		values.Set("cookies", p.Cookies)
	}
	if p.Country != "" {
		values.Set("country", p.Country)
	}
	if p.WaitMS > 0 {
		values.Set("wait_ms", strconv.Itoa(p.WaitMS))
	}
	if p.WaitSelector != "" {
		values.Set("wait_selector", p.WaitSelector)
	}
	if p.FollowRedirects != nil {
		values.Set("follow_redirects", strconv.FormatBool(*p.FollowRedirects))
	}
	if p.BlockResources != nil {
		values.Set("block_resources", strconv.FormatBool(*p.BlockResources))
	}
	if p.Clear != nil {
		values.Set("clear", strconv.FormatBool(*p.Clear))
	}
	if p.AutoRuns != nil {
		values.Set("auto_runs", strconv.Itoa(*p.AutoRuns))
	}

	var out UnblockerResult
	if err := s.d.DoFormURLEncoded(ctx, s.d.WebUnblockerURL(), "/request", values, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Countries lists Web Unblocker countries (POST /v1/proxy/unblocker_area).
func (s *UnblockerService) Countries(ctx context.Context) (*Areas, error) {
	var out Areas
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/unblocker_area", map[string]string{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// States lists the states of a country (POST /v1/proxy/unblocker_area_by_country).
// code is required.
func (s *UnblockerService) States(ctx context.Context, code string) (*StateList, error) {
	if err := requireField("code", code); err != nil {
		return nil, err
	}
	var out StateList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/unblocker_area_by_country", map[string]string{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cities lists the cities of a region (POST /v1/proxy/unblocker_city_by_area).
// Both code and region are required.
func (s *UnblockerService) Cities(ctx context.Context, code, region string) (*CityList, error) {
	var missing []string
	if strings.TrimSpace(code) == "" {
		missing = append(missing, "code")
	}
	if strings.TrimSpace(region) == "" {
		missing = append(missing, "region")
	}
	if len(missing) > 0 {
		return nil, &ValidationError{Fields: missing}
	}
	f := map[string]string{"code": code, "region": region}
	var out CityList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/unblocker_city_by_area", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ISPs lists the carriers of a country (POST /v1/proxy/unblocker_city_isp).
// code is required.
func (s *UnblockerService) ISPs(ctx context.Context, code string) (*ISPList, error) {
	if err := requireField("code", code); err != nil {
		return nil, err
	}
	var out ISPList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/unblocker_city_isp", map[string]string{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// --- Browser ---------------------------------------------------------------

// BrowserService exposes the Browser API area and traffic queries on the
// general host.
type BrowserService struct {
	d transport.Doer
}

// BrowserFlowUse is a single Browser API traffic-consumption record.
type BrowserFlowUse struct {
	ID         int   `json:"id"`
	UID        int   `json:"uid"`
	Balance    int64 `json:"balance"`
	AllBuy     int64 `json:"all_buy"`
	Use        int64 `json:"use"`
	Day        int64 `json:"day"`
	ExpireFlow int64 `json:"expire_flow"`
}

// BrowserFlowUseList is the data payload of FlowUse.
type BrowserFlowUseList struct {
	List []BrowserFlowUse `json:"list"`
}

// Countries lists Browser API countries (POST /v1/proxy/browser_area). The API
// does not document a fixed response shape, so the raw data payload is returned.
func (s *BrowserService) Countries(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/browser_area", map[string]string{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FlowUse returns the main account's Browser API traffic consumption over a
// time range (POST /v1/browser_flow/browser_flow_use). Both bounds use the
// "2006-01-02 15:04:05" layout and are required.
func (s *BrowserService) FlowUse(ctx context.Context, start, end string) (*BrowserFlowUseList, error) {
	var missing []string
	if strings.TrimSpace(start) == "" {
		missing = append(missing, "start_time")
	}
	if strings.TrimSpace(end) == "" {
		missing = append(missing, "end_time")
	}
	if len(missing) > 0 {
		return nil, &ValidationError{Fields: missing}
	}
	f := map[string]string{"start_time": start, "end_time": end}
	var out BrowserFlowUseList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/browser_flow/browser_flow_use", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// requireField returns a *ValidationError when val is empty.
func requireField(name, val string) error {
	if strings.TrimSpace(val) == "" {
		return &ValidationError{Fields: []string{name}}
	}
	return nil
}
