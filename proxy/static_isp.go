package proxy

import (
	"context"
	"strconv"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// --- Shared static-product types -------------------------------------------

// StaticIP is a purchased static IP (ISP or dedicated datacenter) as returned
// by the list endpoints.
type StaticIP struct {
	ID         int    `json:"id"`
	IP         string `json:"ip"`
	Port       string `json:"port"`
	Duration   int64  `json:"duration"`
	Region     string `json:"region"`
	ExpireTime int64  `json:"expire_time"`
	CreateTime int64  `json:"create_time"`
	Country    string `json:"country"`
	CountryEN  string `json:"country_en"`
	Status     int    `json:"status"`
	Remark     string `json:"remark"`
	AutoRenew  int    `json:"auto_renew"`
	Type       int    `json:"type"`
	DiffDay    int    `json:"diff_day"`
	Account    string `json:"account"`
	Password   string `json:"password"`
}

// StaticIPList is the data payload of the static IP list endpoints.
type StaticIPList struct {
	List  []StaticIP `json:"list"`
	Page  int        `json:"page"`
	Total int        `json:"total"`
}

// RegionNode is a single selectable region in the region listings.
type RegionNode struct {
	Node   string `json:"node"`
	NodeEN string `json:"node_en"`
	Param  int    `json:"param"`
	Region string `json:"region"`
}

// RegionList is the data payload of the static region endpoints; regions are
// grouped by area name (e.g. "Asia-Pacific").
type RegionList struct {
	List map[string][]RegionNode `json:"list"`
}

// ListStaticParams are the filters for the static IP list endpoints. Page and
// Limit default to 1 and 10 when left at zero.
type ListStaticParams struct {
	Status  string // "" = all, "1" = in use, "2" = expired, "3" = released
	Region  string // optional area code
	KeyWord string // optional keyword (remark, order number, IP)
	// IsAutoRenew filters by auto-renew flag (1 = yes, -1 = no). Nil = no filter.
	IsAutoRenew *int
	Page        int
	Limit       int
}

// ExportStaticParams are the filters for the static IP export endpoints.
type ExportStaticParams struct {
	Status      string
	Region      string
	KeyWord     string
	IsAutoRenew *int
}

// RenewStaticParams renews one or more static IPs. Both fields are required.
type RenewStaticParams struct {
	// IPs is a comma-separated list of IPs, e.g. "1.1.1.1,2.2.2.2".
	IPs string
	// Duration is the activation period: "week" or "month".
	Duration string
}

// RenewSettingParams updates the auto-renewal settings of static IPs. All
// fields are required; the product type is supplied automatically by the
// service method.
type RenewSettingParams struct {
	// IDs is a comma-separated list of IP IDs, e.g. "100,111,112".
	IDs string
	// PackageType is the activation period: "week" or "month".
	PackageType string
	// Status is the renewal status: 1 = normal, -1 = disabled.
	Status int
	// RenewType is the renewal method: 1 = wallet, 2 = credit card.
	RenewType int
}

// StaticISPService covers static ISP proxies (the /v1/static_house/* endpoints).
type StaticISPService struct {
	d transport.Doer
}

// OpenStaticISPParams opens static ISP IPs. All fields are required.
type OpenStaticISPParams struct {
	// IPType is the IP grade: "normal" = standard, "premium" = premium.
	IPType string
	// Region is an "area:num" spec, e.g. "hk:1|us-va:2".
	Region string
	// Duration is the activation period: "week" or "month".
	Duration string
	// Num is the number of IPs to open.
	Num int
}

// Open purchases static ISP IPs (POST /v1/static_house/open).
func (s *StaticISPService) Open(ctx context.Context, p OpenStaticISPParams) error {
	v := &validator{method: "StaticISP.Open"}
	v.str("ip_type", p.IPType)
	v.str("region", p.Region)
	v.str("duration", p.Duration)
	v.nonZero("num", p.Num)
	if err := v.err(); err != nil {
		return err
	}
	f := form{
		"ip_type":  p.IPType,
		"region":   p.Region,
		"duration": p.Duration,
		"num":      strconv.Itoa(p.Num),
	}
	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/static_house/open", f, nil)
}

// List returns purchased static ISP IPs (POST /v1/static_house/list).
func (s *StaticISPService) List(ctx context.Context, p ListStaticParams) (*StaticIPList, error) {
	return listStatic(ctx, s.d, "/v1/static_house/list", p)
}

// Export returns the static ISP IP list as a file stream
// (POST /v1/static_house/export). The bytes are typically CSV.
func (s *StaticISPService) Export(ctx context.Context, p ExportStaticParams) ([]byte, error) {
	return s.d.DoMultipartRaw(ctx, s.d.BaseURL(), "/v1/static_house/export", exportFields(p))
}

// Region lists the available static ISP regions (POST /v1/static_house/region).
// ispType is required: "isp-resi" = static residential, "isp-resi-hq" = premium.
func (s *StaticISPService) Region(ctx context.Context, ispType string) (*RegionList, error) {
	v := &validator{method: "StaticISP.Region"}
	v.str("isp_type", ispType)
	if err := v.err(); err != nil {
		return nil, err
	}
	var out RegionList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/static_house/region", form{"isp_type": ispType}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Renew renews static ISP IPs (POST /v1/static_house/renew).
func (s *StaticISPService) Renew(ctx context.Context, p RenewStaticParams) error {
	return renewStatic(ctx, s.d, "StaticISP.Renew", "/v1/static_house/renew", p)
}

// RenewSetting updates static ISP auto-renewal settings
// (POST /v1/static_house/renew_setting).
func (s *StaticISPService) RenewSetting(ctx context.Context, p RenewSettingParams) error {
	return renewSettingStatic(ctx, s.d, "StaticISP.RenewSetting", "/v1/static_house/renew_setting", "static_house", p)
}

// --- Shared implementations -------------------------------------------------

func listStatic(ctx context.Context, d transport.Doer, path string, p ListStaticParams) (*StaticIPList, error) {
	defaultPage(&p.Page, &p.Limit)
	f := form{}
	f.optStr("status", p.Status)
	f.optStr("region", p.Region)
	f.optStr("key_word", p.KeyWord)
	f.optIntPtr("is_auto_renew", p.IsAutoRenew)
	f.reqInt("page", p.Page)
	f.reqInt("limit", p.Limit)
	var out StaticIPList
	if err := d.DoMultipart(ctx, d.BaseURL(), path, f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func exportFields(p ExportStaticParams) form {
	f := form{}
	f.optStr("status", p.Status)
	f.optStr("region", p.Region)
	f.optStr("key_word", p.KeyWord)
	f.optIntPtr("is_auto_renew", p.IsAutoRenew)
	return f
}

func renewStatic(ctx context.Context, d transport.Doer, method, path string, p RenewStaticParams) error {
	v := &validator{method: method}
	v.str("renew_ip_list", p.IPs)
	v.str("duration", p.Duration)
	if err := v.err(); err != nil {
		return err
	}
	f := form{"renew_ip_list": p.IPs, "duration": p.Duration}
	return d.DoMultipart(ctx, d.BaseURL(), path, f, nil)
}

func renewSettingStatic(ctx context.Context, d transport.Doer, method, path, typeVal string, p RenewSettingParams) error {
	v := &validator{method: method}
	v.str("ids", p.IDs)
	v.str("package_type", p.PackageType)
	v.nonZero("status", p.Status)
	v.nonZero("renew_type", p.RenewType)
	if err := v.err(); err != nil {
		return err
	}
	f := form{
		"type":         typeVal,
		"ids":          p.IDs,
		"package_type": p.PackageType,
		"status":       strconv.Itoa(p.Status),
		"renew_type":   strconv.Itoa(p.RenewType),
	}
	return d.DoMultipart(ctx, d.BaseURL(), path, f, nil)
}
