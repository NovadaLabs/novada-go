// Package proxy implements the Novada proxy management endpoints (the /v1/*
// multipart/form-data APIs). Construct it via the top-level client and reach
// its sub-services through the exported fields, e.g. client.Proxy.Whitelist.
package proxy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// Product identifies a Novada proxy product line. The numeric values are fixed
// by the API; not every product is valid for every endpoint (see each method's
// documentation).
type Product int

// Known product identifiers.
const (
	ProductResidential Product = 1  // Residential
	ProductRotatingISP Product = 2  // Rotating ISP
	ProductRotatingDC  Product = 3  // Rotating Datacenter
	ProductUnlimited   Product = 4  // Unlimited
	ProductStaticISP   Product = 5  // Static ISP
	ProductUnblocker   Product = 7  // Web Unblocker
	ProductMobile      Product = 9  // Mobile
	ProductBrowserAPI  Product = 10 // Browser API
)

// Service is the entry point for proxy management endpoints. Access the
// sub-services through its exported fields.
type Service struct {
	// Account manages proxy sub-accounts (/v1/proxy_account/*).
	Account *AccountService
	// Whitelist manages IP whitelists (/v1/white_list/*).
	Whitelist *WhitelistService
	// Residential covers residential proxy areas and traffic.
	Residential *ResidentialService
	// Mobile covers mobile proxy areas and traffic.
	Mobile *MobileService
	// RotatingISP covers rotating ISP proxy areas and traffic.
	RotatingISP *RotatingISPService
	// RotatingDC covers rotating datacenter proxy areas and traffic.
	RotatingDC *RotatingDCService
	// StaticISP covers static ISP proxies (/v1/static_house/*).
	StaticISP *StaticISPService
	// DedicatedDC covers dedicated datacenter proxies (/v1/static/*).
	DedicatedDC *DedicatedDCService
	// Unlimited covers unlimited proxy servers (/v1/unlimited/*).
	Unlimited *UnlimitedService
	// ProhibitDomain manages blocked domains (/v1/prohibit_domain/*).
	ProhibitDomain *ProhibitDomainService
}

// New constructs a proxy Service backed by d. It is called by the top-level
// client; most users access it via client.Proxy.
func New(d transport.Doer) *Service {
	return &Service{
		Account:        &AccountService{d: d},
		Whitelist:      &WhitelistService{d: d},
		Residential:    &ResidentialService{d: d},
		Mobile:         &MobileService{d: d},
		RotatingISP:    &RotatingISPService{d: d},
		RotatingDC:     &RotatingDCService{d: d},
		StaticISP:      &StaticISPService{d: d},
		DedicatedDC:    &DedicatedDCService{d: d},
		Unlimited:      &UnlimitedService{d: d},
		ProhibitDomain: &ProhibitDomainService{d: d},
	}
}

// ValidationError reports that one or more required parameters were missing,
// detected client-side before any request was sent.
type ValidationError struct {
	// Method is the SDK method that performed the validation.
	Method string
	// Fields lists the names of the missing required parameters.
	Fields []string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("proxy: %s: missing required field(s): %s",
		e.Method, strings.Join(e.Fields, ", "))
}

// form is a small builder for multipart field maps. Optional setters drop
// empty/zero values; required setters always write so the server-side required
// check is satisfied.
type form map[string]string

func (f form) reqStr(key, val string) { f[key] = val }

func (f form) optStr(key, val string) {
	if val != "" {
		f[key] = val
	}
}

func (f form) reqInt(key string, val int) { f[key] = strconv.Itoa(val) }

func (f form) optInt(key string, val int) {
	if val != 0 {
		f[key] = strconv.Itoa(val)
	}
}

func (f form) optIntPtr(key string, val *int) {
	if val != nil {
		f[key] = strconv.Itoa(*val)
	}
}

// validator accumulates missing required fields and turns them into a
// *ValidationError.
type validator struct {
	method  string
	missing []string
}

func (v *validator) str(name, val string) {
	if strings.TrimSpace(val) == "" {
		v.missing = append(v.missing, name)
	}
}

func (v *validator) nonZero(name string, val int) {
	if val == 0 {
		v.missing = append(v.missing, name)
	}
}

func (v *validator) err() error {
	if len(v.missing) == 0 {
		return nil
	}
	return &ValidationError{Method: v.method, Fields: v.missing}
}

// defaultPage applies the API's pagination defaults (page=1, limit=10) when the
// caller leaves them at zero, so list calls work without boilerplate.
func defaultPage(page, limit *int) {
	if *page == 0 {
		*page = 1
	}
	if *limit == 0 {
		*limit = 10
	}
}
