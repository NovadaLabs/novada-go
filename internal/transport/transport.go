// Package transport defines the minimal interface that sub-service packages
// (proxy, scraper, wallet) use to perform HTTP requests. The top-level
// *novada.Client satisfies it, which lets the sub-packages route requests
// without importing the novada package (avoiding an import cycle).
package transport

import (
	"context"
	"net/url"
)

// Doer is the subset of the top-level client used by sub-services. It exposes
// the two encoding helpers plus the three host accessors so a sub-service can
// pick the correct base URL for each endpoint.
type Doer interface {
	// DoMultipart sends a multipart/form-data POST and decodes the response
	// envelope into out.
	DoMultipart(ctx context.Context, baseURL, path string, fields map[string]string, out any) error
	// DoMultipartRaw is like DoMultipart but returns the raw response body,
	// for endpoints that return a file stream instead of the JSON envelope.
	DoMultipartRaw(ctx context.Context, baseURL, path string, fields map[string]string) ([]byte, error)
	// DoFormURLEncoded sends an application/x-www-form-urlencoded POST and
	// decodes the response envelope into out.
	DoFormURLEncoded(ctx context.Context, baseURL, path string, values url.Values, out any) error
	// DoFormURLEncodedRaw is like DoFormURLEncoded but returns the raw response
	// body, for scraper /request responses that are not the JSON envelope.
	DoFormURLEncodedRaw(ctx context.Context, baseURL, path string, values url.Values) ([]byte, error)
	// BaseURL returns the general host for /v1/* endpoints.
	BaseURL() string
	// WebUnblockerURL returns the Web Unblocker host.
	WebUnblockerURL() string
	// ScraperURL returns the Scraper API host.
	ScraperURL() string
}
