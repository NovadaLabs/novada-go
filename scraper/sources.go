package scraper

import "context"

// APIService groups strongly typed Scraper API scrapers (Target =
// TargetScraperAPI). Each method wraps Service.Do with a known scraper_id.
type APIService struct {
	svc *Service

	// YouTube exposes YouTube scrapers.
	YouTube *YouTubeService
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
