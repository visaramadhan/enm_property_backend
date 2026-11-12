package models

import "time"

// Agreement adalah versi perjanjian kemitraan yang harus disetujui mitra.
type Agreement struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Checklist   []string  `json:"checklist"`
	IsActive    bool      `json:"is_active"`
	EffectiveAt time.Time `json:"effective_at"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// AgreementAcceptance adalah bukti persetujuan perjanjian oleh mitra.
type AgreementAcceptance struct {
	ID               string    `json:"id"`
	AgreementID      string    `json:"agreement_id"`
	AgreementVersion string    `json:"agreement_version"`
	UserID           string    `json:"user_id"`
	PartnerID        string    `json:"partner_id"`
	AcceptedAt       time.Time `json:"accepted_at"`
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
}

// Subscription adalah masa aktif paket yang dimiliki mitra.
type Subscription struct {
	ID        string             `json:"id"`
	PartnerID string             `json:"partner_id"`
	PlanID    string             `json:"plan_id"`
	PlanName  string             `json:"plan_name"`
	OrderID   string             `json:"order_id"`
	StartsAt  time.Time          `json:"starts_at"`
	EndsAt    time.Time          `json:"ends_at"`
	Status    SubscriptionStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// IsActiveAt memeriksa apakah langganan masih berlaku pada waktu tertentu.
func (s *Subscription) IsActiveAt(t time.Time) bool {
	return s.Status == SubscriptionStatusActive && !t.After(s.EndsAt)
}

// DaysRemaining menghitung sisa hari masa aktif, dibulatkan ke atas dan tidak negatif.
func (s *Subscription) DaysRemaining(t time.Time) int {
	if t.After(s.EndsAt) {
		return 0
	}
	days := int(s.EndsAt.Sub(t).Hours()/24) + 1
	if days < 0 {
		return 0
	}
	return days
}
