package dto

import "time"

// ContactResponse adalah respons GET /properties/{id}/contact berisi nomor
// WhatsApp mitra, atau nomor admin sebagai cadangan.
type ContactResponse struct {
	PropertyID    string `json:"property_id"`
	PropertyTitle string `json:"property_title"`
	PartnerName   string `json:"partner_name"`
	Phone         string `json:"phone"`
	WhatsAppURL   string `json:"whatsapp_url"`
	Message       string `json:"message"`
	IsFallback    bool   `json:"is_fallback"`
}

// ContactFormRequest adalah body POST /properties/{slug}/contact-form.
//
// Dikirim pengunjung publik yang mengisi form kontak pada halaman detail
// properti. Data tercatat sebagai lead dengan kanal "form".
type ContactFormRequest struct {
	Name    string `json:"name" validate:"required,min=2,max=100"`
	Phone   string `json:"phone" validate:"required,min=6,max=30"`
	Message string `json:"message" validate:"required,min=5,max=1000"`
}

// ContactFormResponse adalah respons setelah form kontak tersimpan.
type ContactFormResponse struct {
	Recorded bool `json:"recorded"`
}

// LeadView adalah satu catatan lead properti yang ditampilkan pada tabel mitra.
type LeadView struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	UserID    string    `json:"user_id,omitempty"`
	UserName  string    `json:"user_name,omitempty"`
	UserEmail string    `json:"user_email,omitempty"`
	Name      string    `json:"name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PartnerPublicView adalah profil pengelola (mitra) beserta katalog miliknya
// untuk halaman publik /pengelola/{id}.
type PartnerPublicView struct {
	ID            string            `json:"id"`
	BusinessName  string            `json:"business_name"`
	PICName       string            `json:"pic_name"`
	City          string            `json:"city"`
	Province      string            `json:"province"`
	JoinedAt      *time.Time        `json:"joined_at,omitempty"`
	TotalListings int               `json:"total_listings"`
	Properties    []PropertySummary `json:"properties"`
}
