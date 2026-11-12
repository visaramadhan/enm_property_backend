package dto

import (
	"time"

	"enmproperty/backend/models"
	"enmproperty/backend/pkg/money"
)

// CheckoutRequest adalah body POST /subscriptions/checkout.
//
// Data mitra hanya diperlukan bila pengguna belum memiliki profil kemitraan.
type CheckoutRequest struct {
	PackageSlug string                  `json:"package_slug" validate:"required,max=60"`
	Mode        string                  `json:"mode" validate:"omitempty,oneof=new renewal upgrade"`
	Partner     *RegisterPartnerRequest `json:"partner"`
	// AgreementID diisi klien dari GET /agreements/active.
	AgreementID    string `json:"agreement_id" validate:"omitempty,max=64"`
	AgreeAgreement bool   `json:"agree_agreement"`
}

// OrderView adalah representasi transaksi langganan untuk status dan invoice.
type OrderView struct {
	ID               string     `json:"id"`
	Code             string     `json:"code"`
	UserID           string     `json:"user_id"`
	PartnerID        string     `json:"partner_id"`
	PartnerName      string     `json:"partner_name"`
	PlanID           string     `json:"plan_id"`
	PlanName         string     `json:"plan_name"`
	PlanPrice        int64      `json:"plan_price"`
	PlanPriceLabel   string     `json:"plan_price_label"`
	PlanDurationDays int        `json:"plan_duration_days"`
	Amount           int64      `json:"amount"`
	AmountLabel      string     `json:"amount_label"`
	Kind             string     `json:"kind"`
	KindLabel        string     `json:"kind_label"`
	Status           string     `json:"status"`
	StatusLabel      string     `json:"status_label"`
	Provider         string     `json:"provider"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// NewOrderView membentuk tampilan order dari entitas domain.
func NewOrderView(order *models.Order) OrderView {
	if order == nil {
		return OrderView{}
	}
	return OrderView{
		ID:               order.ID,
		Code:             order.Code,
		UserID:           order.UserID,
		PartnerID:        order.PartnerID,
		PartnerName:      order.PartnerName,
		PlanID:           order.PlanID,
		PlanName:         order.PlanName,
		PlanPrice:        order.PlanPrice,
		PlanPriceLabel:   money.Rupiah(order.PlanPrice),
		PlanDurationDays: order.PlanDurationDays,
		Amount:           order.Amount,
		AmountLabel:      money.Rupiah(order.Amount),
		Kind:             string(order.Kind),
		KindLabel:        orderKindLabel(order.Kind),
		Status:           string(order.Status),
		StatusLabel:      order.Status.Label(),
		Provider:         string(order.Provider),
		PaidAt:           order.PaidAt,
		ExpiresAt:        order.ExpiresAt,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

// CheckoutResponse adalah respons POST /subscriptions/checkout.
type CheckoutResponse struct {
	Order           OrderView `json:"order"`
	SnapToken       string    `json:"snap_token,omitempty"`
	SnapRedirectURL string    `json:"snap_redirect_url,omitempty"`
	ClientKey       string    `json:"client_key,omitempty"`
	Simulated       bool      `json:"simulated"`
	FreeActive      bool      `json:"free_active"`
}

// OrderStatusView adalah respons GET /subscriptions/orders/{orderCode}/status.
type OrderStatusView struct {
	Order        OrderView         `json:"order"`
	CanRetry     bool              `json:"can_retry"`
	Subscription *SubscriptionView `json:"subscription,omitempty"`
}

// ChargeRequest adalah body POST /subscriptions/orders/{orderCode}/charge.
//
// Method bernilai "qris" atau nama bank virtual account (bca/bni/bri/permata).
type ChargeRequest struct {
	Method string `json:"method" validate:"required,oneof=qris bca bni bri permata"`
}

// PaymentView adalah respons instruksi pembayaran QRIS/VA untuk satu order.
//
// Halaman pembayaran memakai QRString/QRURL untuk QRIS dan Number untuk VA.
type PaymentView struct {
	Order        OrderView  `json:"order"`
	Method       string     `json:"method"`
	MethodLabel  string     `json:"method_label"`
	Channel      string     `json:"channel"`
	ChannelLabel string     `json:"channel_label"`
	Number       string     `json:"number,omitempty"`
	QRURL        string     `json:"qr_url,omitempty"`
	QRString     string     `json:"qr_string,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsPaid       bool       `json:"is_paid"`
	CanRetry     bool       `json:"can_retry"`
}

// NewPaymentView membentuk instruksi pembayaran dari entitas order.
// Flag CanRetry dihitung di lapisan service karena bergantung pada waktu.
func NewPaymentView(order *models.Order) PaymentView {
	if order == nil {
		return PaymentView{}
	}
	return PaymentView{
		Order:        NewOrderView(order),
		Method:       order.PaymentMethod,
		MethodLabel:  paymentMethodLabel(order.PaymentMethod),
		Channel:      order.PaymentChannel,
		ChannelLabel: paymentChannelLabel(order.PaymentChannel),
		Number:       order.PaymentNumber,
		QRURL:        order.PaymentQRURL,
		QRString:     order.PaymentQRString,
		ExpiresAt:    order.PaymentExpiresAt,
		IsPaid:       order.IsPaid(),
	}
}

// SubscriptionView adalah representasi masa aktif paket mitra, termasuk flag
// alert yang dipakai PackageAlertBanner di seluruh halaman mitra.
type SubscriptionView struct {
	ID                 string    `json:"id"`
	PartnerID          string    `json:"partner_id"`
	PlanID             string    `json:"plan_id"`
	PlanName           string    `json:"plan_name"`
	Status             string    `json:"status"`
	StatusLabel        string    `json:"status_label"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	DaysRemaining      int       `json:"days_remaining"`
	IsActive           bool      `json:"is_active"`
	ShowAlert          bool      `json:"show_alert"`
	AlertThresholdDays int       `json:"alert_threshold_days"`
	ListingsUsed       int       `json:"listings_used"`
	ListingQuota       int       `json:"listing_quota"`
	ListingRemaining   int       `json:"listing_remaining"`
	// BoostQuota adalah kuota boost listing paket aktif; 0 berarti tanpa boost.
	BoostQuota int `json:"boost_quota"`
	// BoostUsed adalah jumlah listing mitra yang sedang di-boost bersamaan.
	BoostUsed int `json:"boost_used"`
	// WhatsAppAccess menandakan mitra boleh mengisi nomor WhatsApp sendiri.
	// Paket Gratis bernilai false sehingga kontak memakai nomor sistem.
	WhatsAppAccess bool      `json:"whatsapp_access"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastOrderCode  string    `json:"last_order_code,omitempty"`
}

// NewSubscriptionView membentuk tampilan langganan dari entitas domain.
// Sisa hari, flag alert, dan pemakaian kuota dihitung di lapisan service.
func NewSubscriptionView(subscription *models.Subscription) SubscriptionView {
	if subscription == nil {
		return SubscriptionView{}
	}
	return SubscriptionView{
		ID:          subscription.ID,
		PartnerID:   subscription.PartnerID,
		PlanID:      subscription.PlanID,
		PlanName:    subscription.PlanName,
		Status:      string(subscription.Status),
		StatusLabel: subscriptionStatusLabel(subscription.Status),
		StartsAt:    subscription.StartsAt,
		EndsAt:      subscription.EndsAt,
		IsActive:    subscription.Status == models.SubscriptionStatusActive,
		UpdatedAt:   subscription.UpdatedAt,
	}
}

// InvoiceItem adalah satu baris rincian tagihan pada invoice.
type InvoiceItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Total       int64  `json:"total"`
	TotalLabel  string `json:"total_label"`
}

// InvoiceView adalah respons GET /invoices/{orderCode}.
type InvoiceView struct {
	InvoiceNumber string        `json:"invoice_number"`
	Order         OrderView     `json:"order"`
	Partner       PartnerDetail `json:"partner"`
	Items         []InvoiceItem `json:"items"`
	Subtotal      int64         `json:"subtotal"`
	SubtotalLabel string        `json:"subtotal_label"`
	Total         int64         `json:"total"`
	TotalLabel    string        `json:"total_label"`
	IssuedAt      time.Time     `json:"issued_at"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	IsPaid        bool          `json:"is_paid"`
}

// PartnerDetail adalah data mitra pada invoice.
type PartnerDetail struct {
	ID           string `json:"id"`
	BusinessName string `json:"business_name"`
	PICName      string `json:"pic_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Province     string `json:"province"`
}

// MidtransNotification adalah payload webhook Midtrans yang diproses handler.
type MidtransNotification struct {
	OrderID           string `json:"order_id" validate:"required"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
}

// orderKindLabel mengembalikan label jenis order dalam bahasa Indonesia.
func orderKindLabel(kind models.OrderKind) string {
	switch kind {
	case models.OrderKindNew:
		return "Pembelian baru"
	case models.OrderKindRenewal:
		return "Perpanjangan"
	case models.OrderKindUpgrade:
		return "Peningkatan paket"
	case models.OrderKindFree:
		return "Paket gratis"
	default:
		return string(kind)
	}
}

// subscriptionStatusLabel mengembalikan label status langganan dalam bahasa Indonesia.
func subscriptionStatusLabel(status models.SubscriptionStatus) string {
	switch status {
	case models.SubscriptionStatusActive:
		return "Aktif"
	case models.SubscriptionStatusExpired:
		return "Kedaluwarsa"
	case models.SubscriptionStatusCancelled:
		return "Dibatalkan"
	default:
		return string(status)
	}
}

// paymentMethodLabel mengembalikan label metode pembayaran.
func paymentMethodLabel(method string) string {
	switch method {
	case "qris":
		return "QRIS"
	case "va":
		return "Virtual Account"
	default:
		return method
	}
}

// paymentChannelLabel mengembalikan label kanal pembayaran QRIS/VA.
func paymentChannelLabel(channel string) string {
	switch channel {
	case "qris":
		return "QRIS"
	case "bca":
		return "BCA Virtual Account"
	case "bni":
		return "BNI Virtual Account"
	case "bri":
		return "BRI Virtual Account"
	case "permata":
		return "Permata Virtual Account"
	default:
		return channel
	}
}
