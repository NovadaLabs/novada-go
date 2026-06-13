// Package scraper implements the Novada scraping endpoints. Scrape jobs are
// submitted to a single POST /request endpoint (application/x-www-form-urlencoded)
// and distinguished by their scraper_id; the host depends on the Target. The
// query endpoints (areas, balance, unit price) are regular /v1/* APIs on the
// general host. Reach the sub-services via client.Scraper.
package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// Target selects which host a scrape /request is sent to.
type Target int

const (
	// TargetScraperAPI routes to the Scraper API host (scraper.novada.com).
	TargetScraperAPI Target = iota
	// TargetWebUnblocker routes to the Web Unblocker host (webunlocker.novada.com).
	TargetWebUnblocker
)

// Request is a generic scrape job. It covers every scraper_id; the strongly
// typed helpers under Service.API build a Request for known scrapers.
type Request struct {
	// Target selects the host (Scraper API or Web Unblocker).
	Target Target
	// ScraperName is the site name, e.g. "youtube.com". Required.
	ScraperName string
	// ScraperID identifies the scrape operation, e.g.
	// "youtube_video-post_explore". Required.
	ScraperID string
	// Params is the JSON array of per-item parameters; it is marshaled into the
	// scraper_params form field. Required (at least one item).
	Params []map[string]any
	// ReturnErrors sets scraper_errors=true so the response includes scrape
	// errors.
	ReturnErrors bool
}

// Response holds a raw scrape result. The body format depends on the scraper
// (JSON, CSV or XLSX), so it is returned verbatim.
type Response struct {
	// Raw is the unparsed response body.
	Raw string
}

// ValidationError reports required scrape parameters that were missing before a
// request was sent.
type ValidationError struct {
	Fields []string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("scraper: missing required field(s): %s", strings.Join(e.Fields, ", "))
}

// Service is the entry point for scraping endpoints. Access the sub-services
// through its exported fields.
type Service struct {
	d transport.Doer

	// API exposes strongly typed Scraper API scrapers.
	API *APIService
	// Unblocker exposes Web Unblocker scraping and area queries.
	Unblocker *UnblockerService
	// Universal exposes the shared balance/logs/unit-price queries.
	Universal *UniversalService
	// Browser exposes the Browser API area and traffic queries.
	Browser *BrowserService
}

// New constructs a scraper Service backed by d. It is called by the top-level
// client; most users access it via client.Scraper.
func New(d transport.Doer) *Service {
	s := &Service{d: d}
	s.API = &APIService{svc: s}
	s.API.YouTube = &YouTubeService{svc: s}
	s.Unblocker = &UnblockerService{d: d, svc: s}
	s.Universal = &UniversalService{d: d}
	s.Browser = &BrowserService{d: d}
	return s
}

// Do submits a generic scrape job. It marshals Params into scraper_params,
// URL-encodes the form, and routes to the host selected by req.Target.
func (s *Service) Do(ctx context.Context, req Request) (*Response, error) {
	var missing []string
	if strings.TrimSpace(req.ScraperName) == "" {
		missing = append(missing, "scraper_name")
	}
	if strings.TrimSpace(req.ScraperID) == "" {
		missing = append(missing, "scraper_id")
	}
	if len(req.Params) == 0 {
		missing = append(missing, "scraper_params")
	}
	if len(missing) > 0 {
		return nil, &ValidationError{Fields: missing}
	}

	encodedParams, err := json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf("scraper: encode params: %w", err)
	}

	values := url.Values{}
	values.Set("scraper_name", req.ScraperName)
	values.Set("scraper_id", req.ScraperID)
	values.Set("scraper_params", string(encodedParams))
	if req.ReturnErrors {
		values.Set("scraper_errors", "true")
	}

	host := s.d.ScraperURL()
	if req.Target == TargetWebUnblocker {
		host = s.d.WebUnblockerURL()
	}

	body, err := s.d.DoFormURLEncodedRaw(ctx, host, "/request", values)
	if err != nil {
		return nil, err
	}
	return &Response{Raw: string(body)}, nil
}
