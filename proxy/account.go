package proxy

import (
	"context"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// AccountService manages proxy sub-accounts via the /v1/proxy_account/*
// endpoints.
type AccountService struct {
	d transport.Doer
}

// Create adds a proxy sub-account (POST /v1/proxy_account/create). Product,
// Account, Password and Status are required.
func (s *AccountService) Create(ctx context.Context, p CreateAccountParams) error {
	v := &validator{method: "Account.Create"}
	v.nonZero("product", int(p.Product))
	v.str("account", p.Account)
	v.str("password", p.Password)
	v.nonZero("status", p.Status)
	if err := v.err(); err != nil {
		return err
	}

	f := form{}
	f.reqInt("product", int(p.Product))
	f.reqStr("account", p.Account)
	f.reqStr("password", p.Password)
	f.reqInt("status", p.Status)
	f.optStr("remark", p.Remark)
	f.optStr("limit_flow", p.LimitFlow)

	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy_account/create", f, nil)
}

// List returns proxy sub-accounts (POST /v1/proxy_account/list). Product is
// required; Page and Limit default to 1 and 10.
func (s *AccountService) List(ctx context.Context, p ListAccountParams) (*AccountList, error) {
	v := &validator{method: "Account.List"}
	v.nonZero("product", int(p.Product))
	if err := v.err(); err != nil {
		return nil, err
	}
	defaultPage(&p.Page, &p.Limit)

	f := form{}
	f.reqInt("product", int(p.Product))
	f.optInt("status", p.Status)
	f.optStr("account", p.Account)
	f.reqInt("page", p.Page)
	f.reqInt("limit", p.Limit)

	var out AccountList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy_account/list", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a proxy sub-account (POST /v1/proxy_account/update). ID,
// Account and Password are required.
func (s *AccountService) Update(ctx context.Context, p UpdateAccountParams) error {
	v := &validator{method: "Account.Update"}
	v.nonZero("id", p.ID)
	v.str("account", p.Account)
	v.str("password", p.Password)
	if err := v.err(); err != nil {
		return err
	}

	f := form{}
	f.reqInt("id", p.ID)
	f.reqStr("account", p.Account)
	f.reqStr("password", p.Password)
	f.optInt("status", p.Status)
	f.optStr("remark", p.Remark)
	f.optStr("limit_flow", p.LimitFlow)

	return s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy_account/update", f, nil)
}

// ConsumeLog returns per-day traffic consumption for an account
// (POST /v1/proxy_account/consume_log). AccountID is required; Page and Limit
// default to 1 and 10.
func (s *AccountService) ConsumeLog(ctx context.Context, p ConsumeLogParams) (*AccountConsumeLogList, error) {
	v := &validator{method: "Account.ConsumeLog"}
	v.nonZero("account_id", p.AccountID)
	if err := v.err(); err != nil {
		return nil, err
	}
	defaultPage(&p.Page, &p.Limit)

	f := form{}
	f.reqInt("account_id", p.AccountID)
	f.optStr("start_time", p.StartTime)
	f.optStr("end_time", p.EndTime)
	f.reqInt("page", p.Page)
	f.reqInt("limit", p.Limit)

	var out AccountConsumeLogList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/proxy_account/consume_log", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
