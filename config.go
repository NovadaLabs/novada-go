package novada

import (
	"net/http"
	"time"
)

// Default base URLs and tuning values. All three hosts can be overridden via
// the corresponding Option; the defaults target Novada's public production
// endpoints.
const (
	// DefaultBaseURL serves every /v1/* management endpoint (proxy, wallet,
	// and the scraper query endpoints such as unblocker_area or
	// capture/get_balance).
	DefaultBaseURL = "https://api-m.novada.com"
	// DefaultWebUnblockerURL serves the Web Unblocker POST /request endpoint.
	DefaultWebUnblockerURL = "https://webunlocker.novada.com"
	// DefaultScraperURL serves the Scraper API POST /request endpoint.
	DefaultScraperURL = "https://scraper.novada.com"

	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
)

// config holds resolved client settings. It is populated from the defaults and
// then mutated by the supplied Options.
type config struct {
	baseURL         string
	webUnblockerURL string
	scraperURL      string
	httpClient      *http.Client
	timeout         time.Duration
	maxRetries      int
	userAgent       string
}

func defaultConfig() *config {
	return &config{
		baseURL:         DefaultBaseURL,
		webUnblockerURL: DefaultWebUnblockerURL,
		scraperURL:      DefaultScraperURL,
		timeout:         defaultTimeout,
		maxRetries:      defaultMaxRetries,
		userAgent:       userAgent,
	}
}

// Option configures a Client. Options are applied in order by NewClient.
type Option func(*config)

// WithBaseURL overrides the general host used for all /v1/* endpoints.
func WithBaseURL(u string) Option {
	return func(c *config) { c.baseURL = u }
}

// WithWebUnblockerURL overrides the host used for the Web Unblocker
// POST /request endpoint.
func WithWebUnblockerURL(u string) Option {
	return func(c *config) { c.webUnblockerURL = u }
}

// WithScraperURL overrides the host used for the Scraper API POST /request
// endpoint.
func WithScraperURL(u string) Option {
	return func(c *config) { c.scraperURL = u }
}

// WithTimeout sets the per-request timeout. It is ignored when a custom
// *http.Client is supplied via WithHTTPClient.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// WithMaxRetries sets how many times a request is retried on network errors or
// HTTP 429/5xx responses. Business failures (code != 0) are never retried.
func WithMaxRetries(n int) Option {
	return func(c *config) { c.maxRetries = n }
}

// WithHTTPClient supplies a custom *http.Client. When set, WithTimeout is
// ignored and the caller is responsible for configuring timeouts.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) { c.httpClient = hc }
}

// WithUserAgent overrides the User-Agent header sent on every request.
func WithUserAgent(ua string) Option {
	return func(c *config) { c.userAgent = ua }
}
