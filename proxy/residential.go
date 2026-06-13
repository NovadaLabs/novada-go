package proxy

import (
	"context"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// ResidentialService covers residential proxy area listings and traffic
// (the /v1/proxy/* and /v1/residential_flow/* endpoints).
type ResidentialService struct {
	d transport.Doer
}

// CityParams selects a city listing by country code and region/state. Both
// fields are required.
type CityParams struct {
	Code   string // required; country code, e.g. "us"
	Region string // required; region/state name, e.g. "alabama"
}

// ResidentialAreas is the country listing returned by Countries. Continent maps
// continent codes to names; Country groups countries by continent.
type ResidentialAreas struct {
	Continent map[string]string           `json:"continent"`
	Country   []ResidentialContinentGroup `json:"country"`
}

// ResidentialContinentGroup is a continent with its list of countries.
type ResidentialContinentGroup struct {
	Continent     string               `json:"continent"`
	ContinentCode int                  `json:"continent_code"`
	List          []ResidentialCountry `json:"list"`
}

// ResidentialCountry is a single residential proxy country.
type ResidentialCountry struct {
	Code      string `json:"code"`
	Continent int    `json:"continent"`
	IPNum     int    `json:"ip_num"`
	Name      string `json:"name"`
	NameEN    string `json:"name_en"`
	Protocol  int    `json:"protocol"`
}

// ResidentialState is a state/region within a country.
type ResidentialState struct {
	State string `json:"state"`
}

// ResidentialStateList is the data payload of States.
type ResidentialStateList struct {
	Data []ResidentialState `json:"data"`
}

// ResidentialCity is a city within a region.
type ResidentialCity struct {
	Code string `json:"code"`
}

// ResidentialCityList is the data payload of Cities.
type ResidentialCityList struct {
	Data []ResidentialCity `json:"data"`
}

// ResidentialISP is an ISP available in a country.
type ResidentialISP struct {
	ASN      string `json:"asn"`
	ShowName string `json:"show_name"`
}

// ResidentialISPList is the data payload of ISPs.
type ResidentialISPList struct {
	List []ResidentialISP `json:"list"`
}

// Countries lists residential proxy countries grouped by continent
// (POST /v1/proxy/domestic_dynamic_area).
func (s *ResidentialService) Countries(ctx context.Context) (*ResidentialAreas, error) {
	var out ResidentialAreas
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/domestic_dynamic_area", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// States lists the states/regions of a country (POST /v1/proxy/city_by_code).
// code is the country code, e.g. "us".
func (s *ResidentialService) States(ctx context.Context, code string) (*ResidentialStateList, error) {
	v := &validator{method: "Residential.States"}
	v.str("code", code)
	if err := v.err(); err != nil {
		return nil, err
	}
	var out ResidentialStateList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/city_by_code", form{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cities lists the cities of a country region (POST /v1/proxy/region_by_city).
// Both Code and Region are required.
func (s *ResidentialService) Cities(ctx context.Context, p CityParams) (*ResidentialCityList, error) {
	v := &validator{method: "Residential.Cities"}
	v.str("code", p.Code)
	v.str("region", p.Region)
	if err := v.err(); err != nil {
		return nil, err
	}
	f := form{"code": p.Code, "region": p.Region}
	var out ResidentialCityList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/region_by_city", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ISPs lists the ISPs available in a country (POST /v1/proxy/city_isp). code is
// the country code, e.g. "us".
func (s *ResidentialService) ISPs(ctx context.Context, code string) (*ResidentialISPList, error) {
	v := &validator{method: "Residential.ISPs"}
	v.str("code", code)
	if err := v.err(); err != nil {
		return nil, err
	}
	var out ResidentialISPList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy/city_isp", form{"code": code}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns the remaining residential traffic
// (POST /v1/residential_flow/balance).
func (s *ResidentialService) Balance(ctx context.Context) (*FlowBalance, error) {
	var out FlowBalance
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/residential_flow/balance", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ConsumeLog returns the main account's residential traffic consumption over a
// time range (POST /v1/residential_flow/consume_log). Both bounds are required.
func (s *ResidentialService) ConsumeLog(ctx context.Context, tr TimeRange) (*FlowConsumeLogList, error) {
	return consumeLog(ctx, s.d, "Residential.ConsumeLog", "/v1/residential_flow/consume_log", tr)
}

// consumeLog is the shared implementation for the residential/ISP/DC traffic
// consumption endpoints, which share the same request and response shapes.
func consumeLog(ctx context.Context, d transport.Doer, method, path string, tr TimeRange) (*FlowConsumeLogList, error) {
	v := &validator{method: method}
	v.str("start_time", tr.Start)
	v.str("end_time", tr.End)
	if err := v.err(); err != nil {
		return nil, err
	}
	f := form{"start_time": tr.Start, "end_time": tr.End}
	var out FlowConsumeLogList
	if err := d.DoMultipart(ctx, d.BaseURL(), path, f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
