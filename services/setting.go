package services

import (
	"context"
	"mime/multipart"
	"path"
	"strconv"
	"strings"
	"sync"

	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// maskedValue adalah teks tersamar yang dikirim klien untuk pengaturan rahasia.
// Nilai tersamar tidak boleh menimpa nilai asli saat disimpan kembali.
const maskedValue = "********"

// Jenis gambar pengaturan yang dapat diunggah beserta lebar maksimumnya.
const (
	settingImageKindLogo      = "logo"
	settingImageKindLogin     = "login"
	settingImageLogoMaxWidth  = 1024
	settingImageLoginMaxWidth = 1920
)

// SettingService mengelola pengaturan sistem (BE-14) beserta cache sederhana
// agar modul lain dapat membaca nilai yang sering dipakai tanpa menyentuh
// repository berulang kali.
type SettingService struct {
	deps Deps

	mu    sync.RWMutex
	cache map[string]string
}

// settingDefinition mendeskripsikan satu kunci pengaturan bawaan: label,
// nilai, dan deskripsi bawaannya. Definisi ini dipakai agar halaman
// Pengaturan Sistem selalu menampilkan seluruh bidang meski tabel `settings`
// masih kosong (mis. sebelum seed dijalankan).
type settingDefinition struct {
	Key         string
	Label       string
	Default     string
	Description string
	IsSecret    bool
}

// definitions menyusun daftar kunci pengaturan bawaan secara berurutan.
// Nilai bawaan mengikuti konfigurasi aplikasi bila relevan (nama, nomor
// WhatsApp admin, ambang hari) sehingga tampilan selalu sesuai perilaku nyata.
func (s *SettingService) definitions() []settingDefinition {
	app := s.deps.Config.App
	return []settingDefinition{
		{Key: models.SettingSiteName, Label: "Nama sistem", Default: app.Name, Description: "Nama situs yang tampil di header dan metadata."},
		{Key: models.SettingSiteNavTitle, Label: "Judul di navbar", Default: app.Name, Description: "Teks judul yang tampil pada bilah navigasi."},
		{Key: models.SettingSiteLogoURL, Label: "Logo sistem", Default: "", Description: "URL logo situs. Kosongkan untuk memakai monogram nama situs."},
		{Key: models.SettingBrandColor, Label: "Warna sistem", Default: "", Description: "Warna aksen utama berformat HEX (mis. #1F6F4A). Kosongkan untuk memakai warna bawaan tema."},
		{Key: models.SettingLoginImageURL, Label: "Gambar halaman masuk", Default: "", Description: "URL gambar panel area masuk dan daftar. Kosongkan untuk latar gradien bawaan."},
		{Key: models.SettingAdminWhatsApp, Label: "Nomor WhatsApp admin", Default: app.WhatsAppAdmin, Description: "Nomor WhatsApp admin untuk inquiry mitra."},
		{Key: models.SettingContactFallback, Label: "Nomor WhatsApp cadangan", Default: app.WhatsAppAdmin, Description: "Nomor fallback bila mitra belum mengisi kontak."},
		{Key: models.SettingSubscriptionAlertDay, Label: "Ambang peringatan langganan", Default: strconv.Itoa(app.SubscriptionReminderDays), Description: "Jumlah hari peringatan sebelum langganan berakhir."},
		{Key: models.SettingAgreementEnabled, Label: "Wajib persetujuan perjanjian", Default: "true", Description: "Isi \"true\" bila perjanjian wajib disetujui mitra sebelum mengunggah listing."},
	}
}

// GetAll mengembalikan seluruh pengaturan dengan nilai rahasia tersamar.
//
// Seluruh kunci bawaan selalu disertakan (memakai nilai basis data bila ada,
// atau nilai bawaan bila belum tersimpan) sehingga halaman Pengaturan Sistem
// tidak pernah kosong. Kunci tambahan di luar daftar bawaan tetap ditampilkan
// agar tidak hilang.
func (s *SettingService) GetAll() []dto.SettingView {
	stored := s.deps.Repos.Settings.All()
	byKey := make(map[string]*models.Setting, len(stored))
	for _, item := range stored {
		if item != nil {
			byKey[item.Key] = item
		}
	}

	defs := s.definitions()
	views := make([]dto.SettingView, 0, len(defs))
	known := make(map[string]struct{}, len(defs))
	for _, def := range defs {
		known[def.Key] = struct{}{}

		value := def.Default
		view := dto.SettingView{
			Key:         def.Key,
			Label:       def.Label,
			Description: def.Description,
			IsSecret:    def.IsSecret,
		}
		if item := byKey[def.Key]; item != nil {
			view.UpdatedAt = item.UpdatedAt
			if item.Value != "" {
				value = item.Value
			}
		}
		if def.IsSecret && value != "" {
			value = maskedValue
		}
		view.Value = value
		views = append(views, view)
	}

	for _, item := range stored {
		if item == nil {
			continue
		}
		if _, ok := known[item.Key]; ok {
			continue
		}
		views = append(views, dto.NewSettingView(item))
	}
	return views
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

// SiteNavTitle mengembalikan judul yang tampil pada bilah navigasi. Bila belum
// diisi, nilainya mengikuti nama situs.
func (s *SettingService) SiteNavTitle() string {
	return firstNonEmpty(s.GetValue(models.SettingSiteNavTitle), s.SiteName())
}

// BrandColor mengembalikan warna aksen sistem berformat HEX. Nilai kosong
// berarti antarmuka memakai warna bawaan tema.
func (s *SettingService) BrandColor() string {
	return strings.TrimSpace(s.GetValue(models.SettingBrandColor))
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

// UploadImage menyimpan berkas gambar pengaturan (logo situs atau gambar area
// masuk), mengubahnya menjadi WebP, lalu memperbarui nilai pengaturan terkait
// dan cache-nya. `kind` bernilai "logo" atau "login".
func (s *SettingService) UploadImage(ctx context.Context, kind string, file *multipart.FileHeader, actorID string) (dto.SettingImageResponse, error) {
	var (
		key      string
		maxWidth int
	)
	switch strings.TrimSpace(kind) {
	case settingImageKindLogo:
		key = models.SettingSiteLogoURL
		maxWidth = settingImageLogoMaxWidth
	case settingImageKindLogin:
		key = models.SettingLoginImageURL
		maxWidth = settingImageLoginMaxWidth
	default:
		return dto.SettingImageResponse{}, httpx.BadRequest("Jenis gambar pengaturan tidak dikenal")
	}

	if err := validateImageUpload(s.deps.Config.Storage, file); err != nil {
		return dto.SettingImageResponse{}, err
	}
	raw, err := readUpload(file)
	if err != nil {
		return dto.SettingImageResponse{}, err
	}

	processed, width, height, err := processLandingImage(raw, maxWidth)
	if err != nil {
		return dto.SettingImageResponse{}, err
	}

	objectPath := path.Join("settings", kind+"-"+newID()+".webp")
	url, err := s.deps.Storage.Save(ctx, objectPath, processed)
	if err != nil {
		return dto.SettingImageResponse{}, httpx.Internal(err)
	}

	setting := &models.Setting{Key: key, Value: url, UpdatedBy: actorID, UpdatedAt: timeNow()}
	if err := s.deps.Repos.Settings.Set(setting); err != nil {
		return dto.SettingImageResponse{}, httpx.Internal(err)
	}
	s.storeCache(key, url)

	return dto.SettingImageResponse{
		Key:       key,
		URL:       url,
		Width:     width,
		Height:    height,
		SizeBytes: int64(len(processed)),
	}, nil
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
