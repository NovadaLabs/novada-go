package proxy

import (
	"context"
	"encoding/json"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// MobileService covers mobile proxy areas and traffic (the /v1/proxy/mobile_area
// and /v1/mobile_flow/* endpoints).
type MobileService struct {
	d transport.Doer
}

// MobileUseParams selects a mobile traffic window. Start, End and Granularity
// are required.
type MobileUseParams struct {
	Start string // required; "2006-01-02 15:04:05"
	End   string // required; "2006-01-02 15:04:05"
	// Granularity selects the bucket size: "1"=hour, "2"=day.
	Granularity string
}

// MobileBalance is the remaining mobile traffic in bytes.
type MobileBalance struct {
	Balance int64 `json:"balance"`
}

// Countries lists mobile proxy countries (POST /v1/proxy/mobile_area). The API
// does not document a fixed response shape, so the raw data payload is
// returned for the caller to decode.
func (s *MobileService) Countries(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/mobile_area", form{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Balance returns the remaining mobile traffic
// (POST /v1/mobile_flow/mobile_flow_balance).
func (s *MobileService) Balance(ctx context.Context) (*MobileBalance, error) {
	var out MobileBalance
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/mobile_flow/mobile_flow_balance", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConsumeLog returns the main account's mobile traffic consumption
// (POST /v1/mobile_flow/mobile_flow_use). Start, End and Granularity are
// required.
func (s *MobileService) ConsumeLog(ctx context.Context, p MobileUseParams) (*FlowConsumeLogList, error) {
	v := &validator{method: "Mobile.ConsumeLog"}
	v.str("start_time", p.Start)
	v.str("end_time", p.End)
	v.str("day_or_hour", p.Granularity)
	if err := v.err(); err != nil {
		return nil, err
	}
	f := form{
		"start_time":  p.Start,
		"end_time":    p.End,
		"day_or_hour": p.Granularity,
	}
	var out FlowConsumeLogList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/mobile_flow/mobile_flow_use", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
