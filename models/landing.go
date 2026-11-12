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

	// Judul dan deskripsi tiap bagian (section) landing page. Nilai kosong
	// berarti halaman memakai teks bawaan yang tertanam di komponen.
	LandingKeunggulanTitle    = "keunggulan_title"
	LandingKeunggulanSubtitle = "keunggulan_subtitle"
	LandingPropertiTitle      = "properti_title"
	LandingPropertiSubtitle   = "properti_subtitle"
	LandingPaketTitle         = "paket_title"
	LandingPaketSubtitle      = "paket_subtitle"
	LandingPanduanTitle       = "panduan_title"
	LandingPanduanSubtitle    = "panduan_subtitle"
	LandingFAQTitle           = "faq_title"
	LandingFAQSubtitle        = "faq_subtitle"
)

// LandingContent adalah satu pasangan kunci-nilai konten landing page.
type LandingContent struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
