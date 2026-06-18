# novada-go

[![CI](https://github.com/NovadaLabs/novada-go/actions/workflows/ci.yml/badge.svg)](https://github.com/NovadaLabs/novada-go/actions/workflows/ci.yml)

Go SDK for the [Novada](https://novada.com) API — proxy management, scraping, and wallet endpoints. Standard library only, Go 1.21+.

## Install

```sh
go get github.com/NovadaLabs/novada-go
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	novada "github.com/NovadaLabs/novada-go"
	"github.com/NovadaLabs/novada-go/proxy"
)

func main() {
	client, err := novada.NewClient("YOUR_API_KEY") // or "" to read NOVADA_API_KEY
	if err != nil {
		log.Fatal(err)
	}

	list, err := client.Proxy.Whitelist.List(context.Background(), proxy.ListWhitelistParams{
		Product: proxy.ProductResidential,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("whitelist total=%d\n", list.Total)
}
```

## The three base URLs

Novada serves requests from **three** hosts. All are configurable and default to production:

| Purpose       | Default                          | Used by |
|---------------|----------------------------------|---------|
| General       | `https://api-m.novada.com`       | Every `/v1/*` endpoint (proxy, wallet, and the scraper area/balance/unit queries) |
| Web Unblocker | `https://webunlocker.novada.com` | `Scraper.Unblocker.Scrape` (typed), or `Scraper.Do` with `Target = TargetWebUnblocker` |
| Scraper API   | `https://scraper.novada.com`     | `Scraper.Do` / `Scraper.API.*` with `Target = TargetScraperAPI` |

Only the scrape `POST /request` calls go to the Web Unblocker / Scraper API hosts; everything else (`/v1/*`) uses the general host.

```go
client, _ := novada.NewClient("API_KEY",
	novada.WithBaseURL("https://api-m.novada.com"),
	novada.WithWebUnblockerURL("https://webunlocker.novada.com"),
	novada.WithScraperURL("https://scraper.novada.com"),
	novada.WithTimeout(30*time.Second),
	novada.WithMaxRetries(2),
)
```

The Bearer API key is injected on every request. Retries apply only to network errors and HTTP 429/5xx — never to a non-zero business code.

## Proxy

```go
// Sub-accounts
client.Proxy.Account.Create(ctx, proxy.CreateAccountParams{
	Product: proxy.ProductResidential, Account: "account11", Password: "pass11", Status: 1,
})
client.Proxy.Account.List(ctx, proxy.ListAccountParams{Product: proxy.ProductResidential})

// Whitelist
client.Proxy.Whitelist.Add(ctx, proxy.AddWhitelistParams{Product: 1, IP: "10.10.10.1", Remark: "test"})
client.Proxy.Whitelist.List(ctx, proxy.ListWhitelistParams{Product: 1})

// Residential areas & traffic
client.Proxy.Residential.Countries(ctx)
client.Proxy.Residential.Balance(ctx)
client.Proxy.Residential.ConsumeLog(ctx, proxy.TimeRange{
	Start: "2025-01-01 00:00:00", End: "2025-01-31 23:59:59",
})

// Static ISP / dedicated datacenter
client.Proxy.StaticISP.Open(ctx, proxy.OpenStaticISPParams{
	IPType: "normal", Region: "hk:1|us-va:2", Duration: "week", Num: 3,
})
client.Proxy.DedicatedDC.List(ctx, proxy.ListStaticParams{})
```

Sub-services: `Account`, `Whitelist`, `Residential`, `Mobile`, `RotatingISP`, `RotatingDC`, `StaticISP`, `DedicatedDC`, `Unlimited`, `ProhibitDomain`. Required parameters are validated client-side and reported as a `*proxy.ValidationError` before any request is sent.

## Scraper

```go
// Strongly typed (auto-selects the Scraper API host)
res, err := client.Scraper.API.YouTube.VideoPost(ctx, scraper.YouTubeVideoParams{
	URL: "https://www.youtube.com/watch?v=HAwTwmzgNc4",
})

// Generic driver — any scraper_id, choose the host explicitly
res, err = client.Scraper.Do(ctx, scraper.Request{
	Target:      scraper.TargetScraperAPI, // or scraper.TargetWebUnblocker
	ScraperName: "youtube.com",
	ScraperID:   "youtube_video-post_explore",
	Params:      []map[string]any{{"url": "https://www.youtube.com/watch?v=HAwTwmzgNc4"}},
	ReturnErrors: true,
})
fmt.Println(res.Raw) // raw scrape result (JSON/CSV/XLSX, depending on scraper)

// Web Unblocker — typed scrape; returns a structured result, not raw text
unb, err := client.Scraper.Unblocker.Scrape(ctx, scraper.UnblockerParams{
	TargetURL: "https://www.google.com", // required
	Country:   "us",                     // ResponseFormat defaults to "html"
})
fmt.Println(unb.Code, len(unb.HTML), unb.UseBalance)

// Query endpoints on the general host
client.Scraper.Universal.Balance(ctx)   // /v1/capture/get_balance
client.Scraper.Universal.Unit(ctx)      // /v1/capture/unit
client.Scraper.Unblocker.Countries(ctx) // /v1/proxy/unblocker_area
client.Scraper.Browser.Countries(ctx)   // /v1/proxy/browser_area
```

`Scraper.Do` marshals `Params` to JSON, places it in the `scraper_params` form field, URL-encodes the body, and routes to the host selected by `Target`. Scrape responses are returned raw because their format varies by scraper. `Scraper.Unblocker.Scrape` is the dedicated Web Unblocker call: it sends the endpoint's own fields (`target_url`, `response_format`, `js_render`, `country`, `wait_ms`, …) and decodes the JSON envelope into `*scraper.UnblockerResult` (`HTML`, `Code`, `Msg`, `MsgDetail`, `UseBalance`).

## Wallet

```go
client.Wallet.Balance(ctx)
client.Wallet.UsageRecord(ctx, wallet.UsageRecordParams{Page: 1, Limit: 20})
```

## Error handling

Management endpoints return a uniform envelope `{code, data, msg, timestamp}`; **only `code == 0` is success**. A non-zero code or a non-2xx HTTP status becomes an `*novada.APIError`.

```go
list, err := client.Proxy.Whitelist.List(ctx, proxy.ListWhitelistParams{Product: 1})
if err != nil {
	switch {
	case novada.IsAuthError(err):    // HTTP 401/403
		log.Fatal("invalid API key")
	case novada.IsRateLimited(err):  // HTTP 429
		log.Fatal("rate limited")
	default:
		if code, ok := novada.CodeOf(err); ok {
			log.Fatalf("business error code=%d: %v", code, err)
		}
		log.Fatal(err)
	}
}
```

## Examples

Runnable examples live in [`examples/`](examples/): [`proxy`](examples/proxy), [`scraper`](examples/scraper), [`wallet`](examples/wallet). Set `NOVADA_API_KEY` and run e.g. `go run ./examples/proxy`.

## License

[MIT](LICENSE)
