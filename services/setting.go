package services

import (
	"strings"
	"sync"

	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// maskedValue adalah teks tersamar yang dikirim klien untuk pengaturan rahasia.
// Nilai tersamar tidak boleh menimpa nilai asli saat disimpan kembali.
const maskedValue = "********"

// SettingService mengelola pengaturan sistem (BE-14) beserta cache sederhana
// agar modul lain dapat membaca nilai yang sering dipakai tanpa menyentuh
// repository berulang kali.
type SettingService struct {
	deps Deps

	mu    sync.RWMutex
	cache map[string]string
}

// GetAll mengembalikan seluruh pengaturan dengan nilai rahasia tersamar.
func (s *SettingService) GetAll() []dto.SettingView {
	return dto.NewSettingViews(s.deps.Repos.Settings.All())
}

// Update menyimpan perubahan pengaturan lalu menyegarkan cache nilainya.
func (s *SettingService) Update(actorID string, req dto.UpdateSettingsRequest) ([]dto.SettingView, error) {
	now := timeNow()
	for _, item := range req.Settings {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return nil, httpx.BadRequest("Kunci pengaturan tidak boleh kosong")
		}

		value := strings.TrimSpace(item.Value)
		// Nilai tersamar berarti admin tidak mengubah rahasia yang ada.
		if value == maskedValue {
			continue
		}

		setting := &models.Setting{
			Key:       key,
			Value:     value,
			UpdatedBy: actorID,
			UpdatedAt: now,
		}
		if err := s.deps.Repos.Settings.Set(setting); err != nil {
			return nil, httpx.Internal(err)
		}
		s.storeCache(key, value)
	}
	return s.GetAll(), nil
}

// GetValue mengembalikan nilai pengaturan dengan memanfaatkan cache.
func (s *SettingService) GetValue(key string) string {
	if value, ok := s.loadCache(key); ok {
		return value
	}
	setting, err := s.deps.Repos.Settings.Get(key)
	if err != nil || setting == nil {
		return ""
	}
	s.storeCache(key, setting.Value)
	return setting.Value
}

// AlertThresholdDays mengembalikan ambang hari peringatan masa aktif paket.
func (s *SettingService) AlertThresholdDays() int {
	return parseDays(s.GetValue(models.SettingSubscriptionAlertDay), s.deps.Config.App.SubscriptionReminderDays)
}

// AgreementRequired melaporkan apakah perjanjian wajib disetujui mitra.
func (s *SettingService) AgreementRequired() bool {
	raw := s.GetValue(models.SettingAgreementEnabled)
	if raw == "" {
		return true
	}
	return !strings.EqualFold(raw, "false")
}

// AdminWhatsApp mengembalikan nomor WhatsApp admin dengan fallback konfigurasi.
func (s *SettingService) AdminWhatsApp() string {
	return firstNonEmpty(s.GetValue(models.SettingAdminWhatsApp), s.deps.Config.App.WhatsAppAdmin)
}

// ContactFallback mengembalikan nomor kontak cadangan bila mitra tidak memiliki nomor.
func (s *SettingService) ContactFallback() string {
	return firstNonEmpty(s.GetValue(models.SettingContactFallback), s.AdminWhatsApp())
}

// SiteName mengembalikan nama situs yang dipakai header dan metadata.
func (s *SettingService) SiteName() string {
	return firstNonEmpty(s.GetValue(models.SettingSiteName), s.deps.Config.App.Name)
}

// SiteLogoURL mengembalikan URL logo situs. Nilai kosong berarti antarmuka
// memakai monogram bawaan dari nama situs.
func (s *SettingService) SiteLogoURL() string {
	return s.GetValue(models.SettingSiteLogoURL)
}

// LoginImageURL mengembalikan URL gambar panel area autentikasi (masuk dan
// daftar). Nilai kosong berarti panel memakai latar gradien bawaan.
func (s *SettingService) LoginImageURL() string {
	return s.GetValue(models.SettingLoginImageURL)
}

// loadCache membaca nilai cache tanpa menyentuh repository.
func (s *SettingService) loadCache(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache == nil {
		return "", false
	}
	value, ok := s.cache[key]
	return value, ok
}

// storeCache menyimpan nilai ke cache, membuat peta bila belum ada.
func (s *SettingService) storeCache(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cache == nil {
		s.cache = make(map[string]string)
	}
	s.cache[key] = value
}
