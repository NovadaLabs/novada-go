// Package wallet implements the Novada wallet endpoints (/v1/wallet/*). Reach
// it via client.Wallet.
package wallet

import (
	"context"
	"strconv"

	"github.com/NovadaLabs/novada-go/internal/transport"
)

// Service is the entry point for wallet endpoints.
type Service struct {
	d transport.Doer
}

// New constructs a wallet Service backed by d. It is called by the top-level
// client; most users access it via client.Wallet.
func New(d transport.Doer) *Service {
	return &Service{d: d}
}

// Balance is the wallet balance payload.
type Balance struct {
	Balance int64 `json:"balance"`
}

// UsageRecordParams paginates the usage records. Page and Limit default to 1
// and 10 when left at zero.
type UsageRecordParams struct {
	Page  int
	Limit int
}

// UsageRecord is a single wallet usage/order record.
type UsageRecord struct {
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
	ID           int     `json:"id"`
	UID          int     `json:"uid"`
	OrderType    string  `json:"order_type"`
	Type         int     `json:"type"`
	Source       int     `json:"source"`
	OrderID      string  `json:"order_id"`
	Money        float64 `json:"money"`
	PayMoney     float64 `json:"pay_money"`
	PayType      string  `json:"pay_type"`
	PaySN        string  `json:"pay_sn"`
	PayStatus    int     `json:"pay_status"`
	PayCate      int     `json:"pay_cate"`
	Description  string  `json:"description"`
	Status       int     `json:"status"`
	CreateAdmin  string  `json:"create_admin"`
	IsFirstOrder int     `json:"is_first_order"`
	PayTime      int64   `json:"pay_time"`
	ClosedAt     int64   `json:"closed_at"`
}

// UsageRecordList is the data payload of UsageRecord.
type UsageRecordList struct {
	Count int           `json:"count"`
	List  []UsageRecord `json:"list"`
}

// GetBalance returns the wallet balance (POST /v1/wallet/balance).
func (s *Service) Balance(ctx context.Context) (*Balance, error) {
	var out Balance
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/wallet/balance", map[string]string{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UsageRecord returns paginated wallet usage records
// (POST /v1/wallet/usage_record).
func (s *Service) UsageRecord(ctx context.Context, p UsageRecordParams) (*UsageRecordList, error) {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 10
	}
	f := map[string]string{
		"page":  strconv.Itoa(p.Page),
		"limit": strconv.Itoa(p.Limit),
	}
	var out UsageRecordList
	if err := s.d.DoMultipart(ctx, s.d.BaseURL(), "/v1/wallet/usage_record", f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
