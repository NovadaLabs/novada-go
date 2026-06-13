package proxy

import (
	"context"
	"strconv"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// DedicatedDCService covers dedicated datacenter proxies (the /v1/static/*
// endpoints). It shares the StaticIP/RegionList response types with StaticISP.
type DedicatedDCService struct {
	d transport.Doer
}

// OpenDedicatedDCParams opens dedicated datacenter IPs. All fields are required.
type OpenDedicatedDCParams struct {
	// Region is an "area:num" spec, e.g. "hk:1|tw:2".
	Region string
	// Duration is the activation period: "week" or "month".
	Duration string
	// Num is the number of IPs to open.
	Num int
}

// Open purchases dedicated datacenter IPs (POST /v1/static/open).
func (s *DedicatedDCService) Open(ctx context.Context, p OpenDedicatedDCParams) error {
	v := &validator{method: "DedicatedDC.Open"}
	v.str("region", p.Region)
	v.str("duration", p.Duration)
	v.nonZero("num", p.Num)
	if err := v.err(); err != nil {
		return err
	}
	f := form{
		"region":   p.Region,
		"duration": p.Duration,
		"num":      strconv.Itoa(p.Num),
	}
	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/static/open", f, nil)
}

// List returns purchased dedicated datacenter IPs (POST /v1/static/list).
func (s *DedicatedDCService) List(ctx context.Context, p ListStaticParams) (*StaticIPList, error) {
	return listStatic(ctx, s.d, "/v1/static/list", p)
}

// Export returns the dedicated datacenter IP list as a file stream
// (POST /v1/static/export). The bytes are typically CSV.
func (s *DedicatedDCService) Export(ctx context.Context, p ExportStaticParams) ([]byte, error) {
	return s.d.DoMultipartRaw(ctx, s.d.BaseURL(), "/v1/static/export", exportFields(p))
}

// Region lists the available dedicated datacenter regions
// (POST /v1/static/region).
func (s *DedicatedDCService) Region(ctx context.Context) (*RegionList, error) {
	var out RegionList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/static/region", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Renew renews dedicated datacenter IPs (POST /v1/static/renew).
func (s *DedicatedDCService) Renew(ctx context.Context, p RenewStaticParams) error {
	return renewStatic(ctx, s.d, "DedicatedDC.Renew", "/v1/static/renew", p)
}

// RenewSetting updates dedicated datacenter auto-renewal settings
// (POST /v1/static/renew_setting).
func (s *DedicatedDCService) RenewSetting(ctx context.Context, p RenewSettingParams) error {
	return renewSettingStatic(ctx, s.d, "DedicatedDC.RenewSetting", "/v1/static/renew_setting", "static", p)
}
