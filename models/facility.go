package models

import "time"

// Facility adalah fasilitas properti (mis. Carport, Kolam Renang, Keamanan 24
// jam) yang dikelola melalui halaman master data superadmin dan dipilih mitra
// saat mengunggah listing.
type Facility struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
