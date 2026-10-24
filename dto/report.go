 package dto

import (
	"time"
)

// PropertyStats adalah hitungan properti per status.
type PropertyStats struct {
	Total         int `json:"total"`
	Active        int `json:"active"`
	Draft         int `json:"draft"`
	PendingReview int `json:"pending_review"`
	Inactive      int `json:"inactive"`
	Expired       int `json:"expired"`
}

// PartnerStats adalah hitungan mitra per status.
type PartnerStats struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Pending   int `json:"pending"`
	Suspended int `json:"suspended"`
	Expired   int `json:"expired"`
}

// OrderStats adalah hitungan order per status beserta nilai pendapatannya.
type OrderStats struct {
	Total                 int    `json:"total"`
	Pending               int    `json:"pending"`
	Paid                  int    `json:"paid"`
	Failed                int    `json:"failed"`
	Expired               int    `json:"expired"`
	Cancelled             int    `json:"cancelled"`
	Revenue               int64  `json:"revenue"`
	RevenueLabel          string `json:"revenue_label"`
	RevenueThisMonth      int64  `json:"revenue_this_month"`
	RevenueThisMonthLabel string `json:"revenue_this_month_label"`
}

// SubscriptionStats adalah hitungan langganan aktif, hampir berakhir, dan kedaluwarsa.
type SubscriptionStats struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	ExpiringSoon int `json:"expiring_soon"`
	Expired      int `json:"expired"`
	Cancelled    int `json:"cancelled"`
}

// AdminDashboard adalah respons GET /admin/dashboard.
type AdminDashboard struct {
	Partners         PartnerStats      `json:"partners"`
	Properties       PropertyStats     `json:"properties"`
	Subscriptions    SubscriptionStats `json:"subscriptions"`
	Orders           OrderStats        `json:"orders"`
	PropertyViews    int               `json:"property_views"`
	PropertyContacts int               `json:"property_contacts"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

// OwnerDashboard adalah respons GET /owner/dashboard.
type OwnerDashboard struct {
	Partners      PartnerStats       `json:"partners"`
	Properties    PropertyStats      `json:"properties"`
	Subscriptions SubscriptionStats  `json:"subscriptions"`
	Orders        OrderStats         `json:"orders"`
	TopPartners   []PartnerReportRow `json:"top_partners"`
	GeneratedAt   time.Time          `json:"generated_at"`
}

// VendorDashboard adalah respons GET /vendor/dashboard.
type VendorDashboard struct {
	PartnerID         string            `json:"partner_id"`
	PartnerName       string            `json:"partner_name"`
	PartnerStatus     string            `json:"partner_status"`
	Properties        PropertyStats     `json:"properties"`
	PropertyViews     int               `json:"property_views"`
	PropertyContacts  int               `json:"property_contacts"`
	Favorites         int               `json:"favorites"`
	ListingsQuota     int               `json:"listing_quota"`
	ListingsUsed      int               `json:"listing_used"`
	ListingsRemaining int               `json:"listing_remaining"`
	Subscription      *SubscriptionView `json:"subscription,omitempty"`
	RecentProperties  []PropertySummary `json:"recent_properties"`
	GeneratedAt       time.Time         `json:"generated_at"`
}

// SubscriptionReportRow adalah satu baris laporan langganan.
type SubscriptionReportRow struct {
	OrderCode   string     `json:"order_code"`
	PartnerID   string     `json:"partner_id"`
	PartnerName string     `json:"partner_name"`
	PlanID      string     `json:"plan_id"`
	PlanName    string     `json:"plan_name"`
	Kind        string     `json:"kind"`
	Status      string     `json:"status"`
	StatusLabel string     `json:"status_label"`
	Provider    string     `json:"provider"`
	Amount      int64      `json:"amount"`
	AmountLabel string     `json:"amount_label"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

// SubscriptionReport adalah respons laporan langganan beserta ringkasannya.
type SubscriptionReport struct {
	Rows             []SubscriptionReportRow `json:"rows"`
	Total            int                     `json:"total"`
	TotalAmount      int64                   `json:"total_amount"`
	TotalAmountLabel string                  `json:"total_amount_label"`
}

// PartnerReportRow adalah satu baris laporan mitra dan propertinya.
type PartnerReportRow struct {
	PartnerID          string     `json:"partner_id"`
	BusinessName       string     `json:"business_name"`
	PICName            string     `json:"pic_name"`
	Phone              string     `json:"phone"`
	City               string     `json:"city"`
	Province           string     `json:"province"`
	Status             string     `json:"status"`
	StatusLabel        string     `json:"status_label"`
	PlanName           string     `json:"plan_name"`
	SubscriptionEndsAt *time.Time `json:"subscription_ends_at,omitempty"`
	PropertiesTotal    int        `json:"properties_total"`
	PropertiesActive   int        `json:"properties_active"`
	PropertyViews      int        `json:"property_views"`
	PropertyContacts   int        `json:"property_contacts"`
	JoinedAt           *time.Time `json:"joined_at,omitempty"`
}

// ReportFilterQuery adalah parameter filter laporan admin, owner, dan ekspor.
//
// From dan To memakai format RFC3339 agar sejalan dengan filter repository.
type ReportFilterQuery struct {
	Page      int    `form:"page"`
	PerPage   int    `form:"per_page"`
	Search    string `form:"search"`
	Status    string `form:"status"`
	PlanID    string `form:"plan_id" validate:"omitempty,max=64"`
	PartnerID string `form:"partner_id" validate:"omitempty,max=64"`
	City      string `form:"city" validate:"omitempty,max=100"`
	Province  string `form:"province" validate:"omitempty,max=100"`
	From      string `form:"from" validate:"omitempty"`
	To        string `form:"to" validate:"omitempty"`
	Format    string `form:"format" validate:"omitempty,oneof=xlsx pdf"`
}
