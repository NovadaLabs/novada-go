package proxy

import (
	"context"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// RotatingISPService covers rotating ISP proxy areas and traffic
// (the /v1/proxy/isp_data_area and /v1/isp_flow/* endpoints).
type RotatingISPService struct {
	d transport.Doer
}

// ISPArea is the country listing returned by RotatingISPService.Countries.
type ISPArea struct {
	Country []ISPCountry `json:"country"`
}

// ISPCountry is a single rotating ISP country.
type ISPCountry struct {
	Code string `json:"code"`
	Node string `json:"node"`
}

// Countries lists rotating ISP countries (POST /v1/proxy/isp_data_area).
func (s *RotatingISPService) Countries(ctx context.Context) (*ISPArea, error) {
	var out ISPArea
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/isp_data_area", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns the remaining rotating ISP traffic (POST /v1/isp_flow/balance).
func (s *RotatingISPService) Balance(ctx context.Context) (*FlowBalance, error) {
	var out FlowBalance
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/isp_flow/balance", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConsumeLog returns the main account's rotating ISP traffic consumption over a
// time range (POST /v1/isp_flow/consume_log).
func (s *RotatingISPService) ConsumeLog(ctx context.Context, tr TimeRange) (*FlowConsumeLogList, error) {
	return consumeLog(ctx, s.d, "RotatingISP.ConsumeLog", "/v1/isp_flow/consume_log", tr)
}

// RotatingDCService covers rotating datacenter proxy areas and traffic
// (the /v1/proxy/dynamic_data_area and /v1/dc_flow/* endpoints).
type RotatingDCService struct {
	d transport.Doer
}

// DCArea is the country listing returned by RotatingDCService.Countries.
type DCArea struct {
	Country []DCCountry `json:"country"`
}

// DCCountry is a single rotating datacenter country.
type DCCountry struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Status     int    `json:"status"`
	Continents int    `json:"continents"`
	AreaCode   int    `json:"area_code"`
	NameEN     string `json:"name_en"`
	Available  int    `json:"available"`
	City       DCCity `json:"city"`
}

// DCCity is the city descriptor nested in a DCCountry.
type DCCity struct {
	Code string `json:"code"`
}

// Countries lists rotating datacenter countries (POST /v1/proxy/dynamic_data_area).
func (s *RotatingDCService) Countries(ctx context.Context) (*DCArea, error) {
	var out DCArea
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/dynamic_data_area", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns the remaining rotating datacenter traffic
// (POST /v1/dc_flow/balance).
func (s *RotatingDCService) Balance(ctx context.Context) (*FlowBalance, error) {
	var out FlowBalance
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/dc_flow/balance", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConsumeLog returns the main account's rotating datacenter traffic consumption
// over a time range (POST /v1/dc_flow/consume_log).
func (s *RotatingDCService) ConsumeLog(ctx context.Context, tr TimeRange) (*FlowConsumeLogList, error) {
	return consumeLog(ctx, s.d, "RotatingDC.ConsumeLog", "/v1/dc_flow/consume_log", tr)
}
