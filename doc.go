// Package novada is the Go SDK for the Novada API.
//
// It provides a top-level Client that injects the Bearer API key and routes
// requests to one of three hosts, plus three sub-service packages reached via
// the client:
//
//   - client.Proxy   — proxy management (/v1/* multipart endpoints)
//   - client.Scraper — scraping (/request) and related queries
//   - client.Wallet  — wallet balance and usage records
//
// # Hosts
//
// Three base URLs are used, all overridable via options:
//
//   - General      (https://api-m.novada.com)      — every /v1/* endpoint
//   - Web Unblocker (https://webunlocker.novada.com) — Web Unblocker POST /request
//   - Scraper API   (https://scraper.novada.com)     — Scraper API POST /request
//
// All /v1/* calls go to the general host; only the scrape /request calls are
// routed to the Web Unblocker or Scraper API host (selected per call).
//
// # Responses and errors
//
// Management endpoints return a uniform envelope {code,data,msg,timestamp};
// only code==0 is success. A non-zero code (or a non-2xx HTTP status) becomes
// an *APIError. Use IsAuthError, IsRateLimited and CodeOf to classify errors.
//
// # Quick start
//
//	client, err := novada.NewClient("API_KEY")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	list, err := client.Proxy.Whitelist.List(ctx, proxy.ListWhitelistParams{
//	    Product: proxy.ProductResidential,
//	})
package novada
