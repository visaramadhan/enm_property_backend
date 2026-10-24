package models

import "time"

// Order adalah transaksi pembelian paket langganan. Code dipakai pada URL
// /checkout/status/[orderCode] dan /invoice/[orderCode].
type Order struct {
	ID                    string          `json:"id"`
	Code                  string          `json:"code"`
	UserID                string          `json:"user_id"`
	PartnerID             string          `json:"partner_id"`
	PartnerName           string          `json:"partner_name"`
	PlanID                string          `json:"plan_id"`
	PlanName              string          `json:"plan_name"`
	PlanPrice             int64           `json:"plan_price"`
	PlanDurationDays      int             `json:"plan_duration_days"`
	PlanListingQuota      int             `json:"plan_listing_quota"`
	Amount                int64           `json:"amount"`
	Kind                  OrderKind       `json:"kind"`
	Status                OrderStatus     `json:"status"`
	Provider              PaymentProvider `json:"provider"`
	SnapToken             string          `json:"snap_token,omitempty"`
	SnapRedirectURL       string          `json:"snap_redirect_url,omitempty"`
	MidtransTransactionID string          `json:"midtrans_transaction_id,omitempty"`
	MidtransPaymentType   string          `json:"midtrans_payment_type,omitempty"`
	PaidAt                *time.Time      `json:"paid_at,omitempty"`
	ExpiresAt             time.Time       `json:"expires_at"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// IsPaid memeriksa apakah order sudah dibayar.
func (o *Order) IsPaid() bool { return o.Status == OrderStatusPaid }

// IsSettled memeriksa apakah order sudah final sehingga webhook tidak perlu diproses ulang.
func (o *Order) IsSettled() bool {
	switch o.Status {
	case OrderStatusPaid, OrderStatusExpired, OrderStatusFailed, OrderStatusCancelled:
		return true
	default:
		return false
	}
}
