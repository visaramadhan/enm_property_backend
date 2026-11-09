package dto

import (
	"time"

	"enmproperty/backend/models"
)

// SettingView adalah representasi pengaturan sistem untuk halaman admin.
type SettingView struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	IsSecret    bool      `json:"is_secret"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewSettingView membentuk tampilan pengaturan dengan nilai rahasia tersamar.
func NewSettingView(setting *models.Setting) SettingView {
	if setting == nil {
		return SettingView{}
	}
	return SettingView{
		Key:         setting.Key,
		Value:       setting.MaskedValue(),
		Description: setting.Description,
		IsSecret:    setting.IsSecret,
		UpdatedAt:   setting.UpdatedAt,
	}
}

// NewSettingViews membentuk daftar tampilan pengaturan.
func NewSettingViews(settings []*models.Setting) []SettingView {
	views := make([]SettingView, 0, len(settings))
	for _, setting := range settings {
		views = append(views, NewSettingView(setting))
	}
	return views
}

// SettingUpdate adalah satu pasangan kunci-nilai yang diubah admin.
type SettingUpdate struct {
	Key   string `json:"key" validate:"required,max=60"`
	Value string `json:"value" validate:"max=500"`
}

// UpdateSettingsRequest adalah body PUT /admin/settings.
type UpdateSettingsRequest struct {
	Settings []SettingUpdate `json:"settings" validate:"required,min=1,dive"`
}

// PublicSettingsView adalah pengaturan tampilan yang dibaca halaman publik
// (mis. area masuk) tanpa autentikasi. Hanya berisi nilai yang tidak rahasia.
type PublicSettingsView struct {
	SiteName      string `json:"site_name"`
	LogoURL       string `json:"logo_url"`
	LoginImageURL string `json:"login_image_url"`
}

// SettingImageResponse adalah hasil unggahan gambar pengaturan (logo/latar masuk).
type SettingImageResponse struct {
	Key       string `json:"key"`
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int64  `json:"size_bytes"`
}
