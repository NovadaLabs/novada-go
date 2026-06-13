package novada

import (
	"errors"
	"net/http"
	"os"

	"github.com/NovadaLabs/novada-go/proxy"
	"github.com/NovadaLabs/novada-go/scraper"
	"github.com/NovadaLabs/novada-go/wallet"
)

// ErrNoAPIKey is returned by NewClient when no API key is supplied and the
// NOVADA_API_KEY environment variable is empty.
var ErrNoAPIKey = errors.New("novada: API key is required (pass it to NewClient or set NOVADA_API_KEY)")

// Client is the top-level Novada API client. It injects the Bearer API key on
// every request and routes calls to one of three hosts depending on the
// endpoint. It is safe for concurrent use by multiple goroutines.
//
// Sub-services (proxy, scraper, wallet) live in their own packages and are
// constructed against a *Client.
type Client struct {
	apiKey          string
	baseURL         string
	webUnblockerURL string
	scraperURL      string
	httpClient      *http.Client
	maxRetries      int
	userAgent       string

	// Proxy exposes the proxy management endpoints (/v1/* multipart APIs).
	Proxy *proxy.Service
	// Scraper exposes the scraping endpoints (scrape /request plus queries).
	Scraper *scraper.Service
	// Wallet exposes the wallet endpoints (/v1/wallet/*).
	Wallet *wallet.Service
}

// NewClient creates a Client. The apiKey argument takes precedence; when it is
// empty the NOVADA_API_KEY environment variable is used as a fallback. It
// returns ErrNoAPIKey when neither is set.
//
// All three base URLs default to Novada's public hosts and only need to be
// overridden for private deployments or testing.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("NOVADA_API_KEY")
	}
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.timeout}
	}

	c := &Client{
		apiKey:          apiKey,
		baseURL:         cfg.baseURL,
		webUnblockerURL: cfg.webUnblockerURL,
		scraperURL:      cfg.scraperURL,
		httpClient:      hc,
		maxRetries:      cfg.maxRetries,
		userAgent:       cfg.userAgent,
	}
	c.Proxy = proxy.New(c)
	c.Scraper = scraper.New(c)
	c.Wallet = wallet.New(c)
	return c, nil
}

// BaseURL returns the general host used for /v1/* endpoints.
func (c *Client) BaseURL() string { return c.baseURL }

// WebUnblockerURL returns the host used for the Web Unblocker POST /request
// endpoint.
func (c *Client) WebUnblockerURL() string { return c.webUnblockerURL }

// ScraperURL returns the host used for the Scraper API POST /request endpoint.
func (c *Client) ScraperURL() string { return c.scraperURL }
