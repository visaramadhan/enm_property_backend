package models

import "time"

// Kunci konten landing page (CMS Landing Page).
//
// Konten disimpan sebagai pasangan kunci-nilai pada tabel tersendiri agar tidak
// tercampur dengan pengaturan sistem (BE-14) yang dipakai lintas modul.
const (
	LandingHeroTitle       = "hero_title"
	LandingHeroSubtitle    = "hero_subtitle"
	LandingHeroCTAText     = "hero_cta_text"
	LandingHeroImageURL    = "hero_image_url"
	LandingMetaTitle       = "meta_title"
	LandingMetaDescription = "meta_description"
	LandingMetaOGImage     = "meta_og_image"
	LandingHighlightAreas  = "highlight_areas"
	LandingCTATitle        = "cta_title"
	LandingCTADescription  = "cta_description"
)

// LandingContent adalah satu pasangan kunci-nilai konten landing page.
type LandingContent struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
