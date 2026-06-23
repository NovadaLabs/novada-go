// Command scraper demonstrates the scraping endpoints.
package main

import (
	"context"
	"fmt"
	"log"

	novada "github.com/NovadaLabs/novada-go"
	"github.com/NovadaLabs/novada-go/scraper"
)

func main() {
	client, err := novada.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Generic driver — works for any scraper_id, on either host.
	res, err := client.Scraper.Do(ctx, scraper.Request{
		Target:      scraper.TargetScraperAPI,
		ScraperName: "youtube.com",
		ScraperID:   "youtube_video-post_explore",
		Params: []map[string]any{
			{"url": "https://www.youtube.com/watch?v=HAwTwmzgNc4"},
		},
		ReturnErrors: true,
	})
	if err != nil {
		log.Fatalf("scrape: %v", err)
	}
	fmt.Println(res.Raw)

	// Strongly typed YouTube scraper (routes to the Scraper API host).
	res, err = client.Scraper.API.YouTube.VideoPost(ctx, scraper.YouTubeVideoParams{
		URL: "https://www.youtube.com/watch?v=HAwTwmzgNc4",
	})
	if err != nil {
		log.Fatalf("youtube: %v", err)
	}
	fmt.Println(res.Raw)

	// Strongly typed Google Search scraper (SerpApi; routes to the Scraper API host).
	gs, err := client.Scraper.API.Google.Search(ctx, scraper.GoogleSearchParams{
		Query:   "apple",
		Country: "us",
	})
	if err != nil {
		log.Fatalf("google search: %v", err)
	}
	fmt.Printf("google code=%d cost=%dms results=%d bytes\n",
		gs.Code, gs.CostTime, len(gs.Data.JSON))

	// Web Unblocker — strongly typed scrape (routes to the Web Unblocker host).
	unb, err := client.Scraper.Unblocker.Scrape(ctx, scraper.UnblockerParams{
		TargetURL: "https://www.google.com",
		Country:   "us",
	})
	if err != nil {
		log.Fatalf("unblocker: %v", err)
	}
	fmt.Printf("unblocker code=%d html=%d bytes use_balance=%v\n",
		unb.Code, len(unb.HTML), unb.UseBalance)

	// Query endpoints on the general host.
	bal, err := client.Scraper.Universal.Balance(ctx)
	if err != nil {
		log.Fatalf("balance: %v", err)
	}
	fmt.Printf("scraper balance=%f\n", bal.ScraperBalance)
}
