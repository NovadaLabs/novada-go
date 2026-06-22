package scraper

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// APIService groups strongly typed Scraper API scrapers (Target =
// TargetScraperAPI). Each method wraps Service.Do with a known scraper_id.
type APIService struct {
	svc *Service

	// YouTube exposes YouTube scrapers.
	YouTube *YouTubeService
	// Google exposes Google scrapers (SerpApi).
	Google *GoogleService
}

// YouTubeService groups YouTube scrapers (scraper_name "youtube.com").
type YouTubeService struct {
	svc *Service
}

// init wiring is done lazily here to keep New simple; APIService is created in
// New and its sub-services are attached on first access via accessor methods.

// YouTubeVideoParams are the parameters for YouTubeService.VideoPost.
type YouTubeVideoParams struct {
	// URL is the YouTube video URL. Required.
	URL string
}

// VideoPost scrapes a YouTube video post (scraper_id
// "youtube_video-post_explore"). It is a thin wrapper over Service.Do.
func (s *YouTubeService) VideoPost(ctx context.Context, p YouTubeVideoParams) (*Response, error) {
	return s.svc.Do(ctx, Request{
		Target:      TargetScraperAPI,
		ScraperName: "youtube.com",
		ScraperID:   "youtube_video-post_explore",
		Params:      []map[string]any{{"url": p.URL}},
	})
}

// GoogleService groups Google scrapers (scraper_name "google.com"). Unlike the
// generic Service.Do path, Google scrapers send their parameters as flat
// form fields (not a scraper_params JSON array), so this service builds the
// form directly and decodes the envelope into a structured result.
type GoogleService struct {
	d transport.Doer
}

// GoogleSearchParams configures a Google Search scrape (scraper_id
// "google_search"). Query is required; optional fields are sent only when set.
type GoogleSearchParams struct {
	// Query is the search query or URL. Required (q).
	Query string
	// Device emulates a device type: "desktop", "tablet" or "mobile".
	Device string
	// HTML requests HTML output (json=0) instead of the default JSON output
	// (json=1).
	HTML bool
	// RenderJS enables JavaScript rendering (render_js).
	RenderJS *bool
	// NoCache skips the cache and forces a real-time request (no_cache).
	NoCache *bool
	// AIOverview fetches the AI Overview block; success counts as a response,
	// and a page_token result triggers a second request (two responses total).
	AIOverview *bool
	// Domain is the Google domain to crawl, e.g. "google.com".
	Domain string
	// Country is the two-letter country/region code (gl); server default "us".
	Country string
	// HL is the results language (hl).
	HL string
	// ReturnErrors sets scraper_errors=true so the response includes scrape
	// errors.
	ReturnErrors bool
}

// GoogleSearchResult is the decoded "data" payload of a Google Search scrape.
type GoogleSearchResult struct {
	// Code is the page-level status (200 on success).
	Code int `json:"code"`
	// CostTime is the scrape duration in milliseconds.
	CostTime int `json:"cost_time"`
	// Msg is a short status message.
	Msg string `json:"msg"`
	// Data holds the scraped output.
	Data GoogleSearchData `json:"data"`
}

// GoogleSearchData holds the scraped output. JSON is the raw result array (its
// structure is large and varies by query, so it is left unparsed); HTML is set
// instead when HTML output was requested.
type GoogleSearchData struct {
	Filename string          `json:"filename"`
	HTML     *string         `json:"html"`
	JSON     json.RawMessage `json:"json"`
	TaskID   string          `json:"task_id"`
}

// Search scrapes Google search results (scraper_id "google_search") and decodes
// the structured result. Query is required.
func (s *GoogleService) Search(ctx context.Context, p GoogleSearchParams) (*GoogleSearchResult, error) {
	if err := requireField("q", p.Query); err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("scraper_name", "google.com")
	values.Set("scraper_id", "google_search")
	values.Set("q", p.Query)
	if p.HTML {
		values.Set("json", "0")
	} else {
		values.Set("json", "1")
	}
	if p.Device != "" {
		values.Set("device", p.Device)
	}
	if p.RenderJS != nil {
		values.Set("render_js", strconv.FormatBool(*p.RenderJS))
	}
	if p.NoCache != nil {
		values.Set("no_cache", strconv.FormatBool(*p.NoCache))
	}
	if p.AIOverview != nil {
		values.Set("ai_overview", strconv.FormatBool(*p.AIOverview))
	}
	if p.Domain != "" {
		values.Set("domain", p.Domain)
	}
	if p.Country != "" {
		values.Set("country", p.Country)
	}
	if p.HL != "" {
		values.Set("hl", p.HL)
	}
	if p.ReturnErrors {
		values.Set("scraper_errors", "true")
	}

	var out GoogleSearchResult
	if err := s.d.DoFormURLEncoded(ctx, s.d.ScraperURL(), "/request", values, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
