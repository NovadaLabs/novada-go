package proxy

import (
	"context"
	"strconv"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// UnlimitedService covers unlimited proxy servers (the /v1/unlimited/* endpoints).
type UnlimitedService struct {
	d transport.Doer
}

// UnlimitedHost is a single unlimited proxy server.
type UnlimitedHost struct {
	ID                 int    `json:"id"`
	CreateTime         int64  `json:"create_time"`
	ExpireTime         int64  `json:"expire_time"`
	State              int    `json:"state"`
	Host               string `json:"host"`
	UserRateLimitBytes int64  `json:"user_rate_limit_bytes"`
	OrderID            string `json:"order_id"`
	Days               int    `json:"days"`
	Duration           int    `json:"duration"`
	BandWidth          int64  `json:"band_width"`
	Hardware           string `json:"hardware"`
}

// UnlimitedHostList is the data payload of Hosts.
type UnlimitedHostList struct {
	List  []UnlimitedHost `json:"list"`
	Page  int             `json:"page"`
	Total int             `json:"total"`
}

// Hosts lists unlimited proxy servers (POST /v1/unlimited/host_list). page and
// limit default to 1 and 10 when left at zero.
func (s *UnlimitedService) Hosts(ctx context.Context, page, limit int) (*UnlimitedHostList, error) {
	defaultPage(&page, &limit)
	f := form{
		"page":  strconv.Itoa(page),
		"limit": strconv.Itoa(limit),
	}
	var out UnlimitedHostList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/unlimited/host_list", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
