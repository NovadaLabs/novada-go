package proxy

import (
	"context"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// WhitelistService manages IP whitelists via the /v1/white_list/* endpoints.
type WhitelistService struct {
	d transport.Doer
}

// Add adds an IP to a product's whitelist (POST /v1/white_list/add). Product
// and IP are required. Valid products: 1=Residential, 5=Static ISP,
// 4=Unlimited.
func (s *WhitelistService) Add(ctx context.Context, p AddWhitelistParams) error {
	v := &validator{method: "Whitelist.Add"}
	v.nonZero("product", int(p.Product))
	v.str("ip", p.IP)
	if err := v.err(); err != nil {
		return err
	}

	f := form{}
	f.reqInt("product", int(p.Product))
	f.reqStr("ip", p.IP)
	f.optStr("remark", p.Remark)

	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/white_list/add", f, nil)
}

// List returns whitelisted IPs for a product (POST /v1/white_list/list).
// Product is required; the remaining fields are optional filters.
func (s *WhitelistService) List(ctx context.Context, p ListWhitelistParams) (*WhitelistList, error) {
	v := &validator{method: "Whitelist.List"}
	v.nonZero("product", int(p.Product))
	if err := v.err(); err != nil {
		return nil, err
	}

	f := form{}
	f.reqInt("product", int(p.Product))
	f.optStr("ip", p.IP)
	f.optStr("start_time", p.StartTime)
	f.optStr("end_time", p.EndTime)
	f.optIntPtr("lock", p.Lock)

	var out WhitelistList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/white_list/list", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes one or more IPs from a product's whitelist
// (POST /v1/white_list/del). Product and IPs are required; pass multiple IPs
// comma-separated.
func (s *WhitelistService) Delete(ctx context.Context, p DeleteWhitelistParams) error {
	v := &validator{method: "Whitelist.Delete"}
	v.nonZero("product", int(p.Product))
	v.str("ips", p.IPs)
	if err := v.err(); err != nil {
		return err
	}

	f := form{}
	f.reqInt("product", int(p.Product))
	f.reqStr("ips", p.IPs)

	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/white_list/del", f, nil)
}

// Remark updates the remark of a whitelist entry (POST /v1/white_list/remark).
// Product and ID are required.
func (s *WhitelistService) Remark(ctx context.Context, p RemarkWhitelistParams) error {
	v := &validator{method: "Whitelist.Remark"}
	v.nonZero("product", int(p.Product))
	v.str("id", p.ID)
	if err := v.err(); err != nil {
		return err
	}

	f := form{}
	f.reqInt("product", int(p.Product))
	f.reqStr("id", p.ID)
	f.optStr("remark", p.Remark)

	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/white_list/remark", f, nil)
}
