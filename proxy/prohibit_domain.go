package proxy

import (
	"context"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// ProhibitDomainService manages the blocked-domain list (the
// /v1/prohibit_domain/* endpoints).
type ProhibitDomainService struct {
	d transport.Doer
}

// ProhibitDomain is a single blocked-domain entry.
type ProhibitDomain struct {
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	ID        int    `json:"id"`
	UID       int    `json:"uid"`
	Status    int    `json:"status"`
	Address   string `json:"address"`
}

// ProhibitDomainList is the data payload of List.
type ProhibitDomainList struct {
	List  []ProhibitDomain `json:"list"`
	Total int              `json:"total"`
}

// DeleteProhibitParams identifies which blocked domains to delete. When All is
// true every entry is deleted and ID is ignored; otherwise ID is required.
type DeleteProhibitParams struct {
	ID  string // entry ID; required unless All is true
	All bool   // delete all entries
}

// Add adds a blocked domain (POST /v1/prohibit_domain/add). address is required.
func (s *ProhibitDomainService) Add(ctx context.Context, address string) error {
	v := &validator{method: "ProhibitDomain.Add"}
	v.str("address", address)
	if err := v.err(); err != nil {
		return err
	}
	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/prohibit_domain/add", form{"address": address}, nil)
}

// List returns the blocked-domain list (POST /v1/prohibit_domain/list).
func (s *ProhibitDomainService) List(ctx context.Context) (*ProhibitDomainList, error) {
	var out ProhibitDomainList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/prohibit_domain/list", form{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes blocked domains (POST /v1/prohibit_domain/del). Set All to
// delete every entry, otherwise ID is required.
func (s *ProhibitDomainService) Delete(ctx context.Context, p DeleteProhibitParams) error {
	v := &validator{method: "ProhibitDomain.Delete"}
	if !p.All {
		v.str("id", p.ID)
	}
	if err := v.err(); err != nil {
		return err
	}
	isAll := "2"
	if p.All {
		isAll = "1"
	}
	f := form{"is_all": isAll}
	f.optStr("id", p.ID)
	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/prohibit_domain/del", f, nil)
}
