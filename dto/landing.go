package dto

import (
	"encoding/json"
	"strings"
	"time"

	"enmproperty/backend/models"
)

// LandingContentView adalah konten landing page siap pakai untuk halaman
// publik maupun editor CMS (admin, superadmin, dan owner mode baca).
type LandingContentView struct {
	HeroTitle       string    `json:"hero_title"`
	HeroSubtitle    string    `json:"hero_subtitle"`
	HeroCTAText     string    `json:"hero_cta_text"`
	HeroImageURL    string    `json:"hero_image_url"`
	MetaTitle       string    `json:"meta_title"`
	MetaDescription string    `json:"meta_description"`
	MetaOGImage     string    `json:"meta_og_image"`
	HighlightAreas  []string  `json:"highlight_areas"`
	CTATitle        string    `json:"cta_title"`
	CTADescription  string    `json:"cta_description"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UpdateLandingRequest adalah body PUT /admin/landing.
//
// Batas panjang mengikuti aturan SEO: meta_title 50-60 karakter,
// meta_description 120-160 karakter, hero_title maksimal 70 karakter,
// dan highlight_areas berisi 4-6 area prioritas.
type UpdateLandingRequest struct {
	HeroTitle       string   `json:"hero_title" validate:"required,max=70"`
	HeroSubtitle    string   `json:"hero_subtitle" validate:"max=150"`
	HeroCTAText     string   `json:"hero_cta_text" validate:"max=40"`
	HeroImageURL    string   `json:"hero_image_url" validate:"max=500"`
	MetaTitle       string   `json:"meta_title" validate:"required,min=50,max=60"`
	MetaDescription string   `json:"meta_description" validate:"required,min=120,max=160"`
	MetaOGImage     string   `json:"meta_og_image" validate:"max=500"`
	HighlightAreas  []string `json:"highlight_areas" validate:"required,min=4,max=6,dive,max=60"`
	CTATitle        string   `json:"cta_title" validate:"max=70"`
	CTADescription  string   `json:"cta_description" validate:"max=200"`
}

// LandingImageParam adalah parameter path jenis gambar landing yang diunggah.
type LandingImageParam struct {
	Kind string `uri:"kind" validate:"required,oneof=hero og"`
}

// LandingImageResponse adalah hasil unggah gambar landing setelah dikonversi.
type LandingImageResponse struct {
	URL       string `json:"url"`
	Kind      string `json:"kind"`
	MimeType  string `json:"mime_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int64  `json:"size_bytes"`
}

// LandingPublishResponse adalah hasil pemanggilan webhook revalidasi ISR.
type LandingPublishResponse struct {
	Message     string `json:"message"`
	Revalidated bool   `json:"revalidated"`
}

// NewLandingContentView menyusun tampilan konten landing dari daftar
// pasangan kunci-nilai. Kunci yang belum pernah diisi memakai nilai kosong.
func NewLandingContentView(items []*models.LandingContent) LandingContentView {
	values := make(map[string]string, len(items))
	var updatedAt time.Time
	for _, item := range items {
		if item == nil {
			continue
		}
		values[item.Key] = item.Value
		if item.UpdatedAt.After(updatedAt) {
			updatedAt = item.UpdatedAt
		}
	}

	return LandingContentView{
		HeroTitle:       values[models.LandingHeroTitle],
		HeroSubtitle:    values[models.LandingHeroSubtitle],
		HeroCTAText:     values[models.LandingHeroCTAText],
		HeroImageURL:    values[models.LandingHeroImageURL],
		MetaTitle:       values[models.LandingMetaTitle],
		MetaDescription: values[models.LandingMetaDescription],
		MetaOGImage:     values[models.LandingMetaOGImage],
		HighlightAreas:  DecodeStringList(values[models.LandingHighlightAreas]),
		CTATitle:        values[models.LandingCTATitle],
		CTADescription:  values[models.LandingCTADescription],
		UpdatedAt:       updatedAt,
	}
}

// EncodeStringList menyimpan daftar teks (mis. area prioritas) sebagai JSON
// agar urutannya tetap terjaga di kolom kunci-nilai.
func EncodeStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

// DecodeStringList membaca daftar teks dari nilai JSON. Nilai yang tidak valid
// atau kosong menghasilkan daftar kosong agar halaman tidak gagal dirender.
func DecodeStringList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}
	items := []string{}
	if err := json.Unmarshal([]byte(trimmed), &items); err != nil {
		return []string{}
	}
	return items
}
