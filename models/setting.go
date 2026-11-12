package models

import "time"

// Kunci pengaturan sistem yang dikenal pada modul BE-14.
const (
	SettingSiteName             = "site_name"
	SettingSiteNavTitle         = "site_nav_title"
	SettingSiteLogoURL          = "site_logo_url"
	SettingBrandColor           = "brand_color"
	SettingLoginImageURL        = "login_image_url"
	SettingAdminWhatsApp        = "admin_whatsapp"
	SettingSubscriptionAlertDay = "subscription_alert_days"
	SettingContactFallback      = "contact_fallback_whatsapp"
	SettingAgreementEnabled     = "agreement_required"
)

// Setting adalah pengaturan sistem berbentuk pasangan kunci-nilai.
type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	IsSecret    bool      `json:"is_secret"`
	UpdatedBy   string    `json:"updated_by,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MaskedValue menyamarkan nilai rahasia agar aman dikirim ke admin.
func (s *Setting) MaskedValue() string {
	if s.IsSecret && s.Value != "" {
		return "********"
	}
	return s.Value
}
