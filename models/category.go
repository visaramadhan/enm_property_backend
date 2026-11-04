package models

import "time"

// Category adalah kategori properti yang dipakai untuk pengelompokan katalog
// publik dan dikelola melalui halaman master data superadmin.
type Category struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
