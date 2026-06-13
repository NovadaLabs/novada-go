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

	// Strongly typed YouTube scraper (routes to the Scraper API host).
	res, err := client.Scraper.API.YouTube.VideoPost(ctx, scraper.YouTubeVideoParams{
		URL: "https://www.youtube.com/watch?v=HAwTwmzgNc4",
	})
	if err != nil {
		log.Fatalf("youtube: %v", err)
	}
	fmt.Println(res.Raw)

	// Generic driver — works for any scraper_id, on either host.
	res, err = client.Scraper.Do(ctx, scraper.Request{
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

	// Query endpoints on the general host.
	bal, err := client.Scraper.Universal.Balance(ctx)
	if err != nil {
		log.Fatalf("balance: %v", err)
	}
	fmt.Printf("scraper balance=%d\n", bal.ScraperBalance)
}
