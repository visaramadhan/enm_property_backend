package models

import "time"

// Kunci konten landing page (CMS Landing Page).
//
// Konten disimpan sebagai pasangan kunci-nilai pada tabel tersendiri agar tidak
// tercampur dengan pengaturan sistem (BE-14) yang dipakai lintas modul.
const (
	LandingHeroTitle    = "hero_title"
	LandingHeroSubtitle = "hero_subtitle"
	LandingHeroCTAText  = "hero_cta_text"
	LandingHeroImageURL = "hero_image_url"
	// Daftar URL gambar kolase hero (JSON, maksimal 4). Bila kosong, halaman
	// memakai hero_image_url lalu dilengkapi foto properti teratas.
	LandingHeroImages      = "hero_images"
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
	// Daftar pertanyaan umum yang dapat diedit admin (JSON [{tanya,jawab}]).
	// Bila kosong, halaman memakai teks bawaan komponen.
	LandingFAQItems = "faq_items"

	// Varian bahasa Inggris (English). Hanya berlaku untuk halaman landing
	// versi /en. Nilai kosong berarti tampilan Inggris memakai teks bawaan
	// yang tertanam di komponen, sehingga halaman tidak pernah kosong.
	LandingHeroTitleEN       = "hero_title_en"
	LandingHeroSubtitleEN    = "hero_subtitle_en"
	LandingHeroCTATextEN     = "hero_cta_text_en"
	LandingMetaTitleEN       = "meta_title_en"
	LandingMetaDescriptionEN = "meta_description_en"
	LandingCTATitleEN        = "cta_title_en"
	LandingCTADescriptionEN  = "cta_description_en"

	LandingKeunggulanTitleEN    = "keunggulan_title_en"
	LandingKeunggulanSubtitleEN = "keunggulan_subtitle_en"
	LandingPropertiTitleEN      = "properti_title_en"
	LandingPropertiSubtitleEN   = "properti_subtitle_en"
	LandingPaketTitleEN         = "paket_title_en"
	LandingPaketSubtitleEN      = "paket_subtitle_en"
	LandingPanduanTitleEN       = "panduan_title_en"
	LandingPanduanSubtitleEN    = "panduan_subtitle_en"
	LandingFAQTitleEN           = "faq_title_en"
	LandingFAQSubtitleEN        = "faq_subtitle_en"
	// Varian Inggris daftar pertanyaan umum (JSON [{tanya,jawab}]).
	LandingFAQItemsEN = "faq_items_en"
)

// LandingContent adalah satu pasangan kunci-nilai konten landing page.
type LandingContent struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
