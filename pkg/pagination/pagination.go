// Package pagination menyatukan aturan paginasi agar seluruh endpoint konsisten.
package pagination

import "strconv"

// Batas paginasi yang dipakai seluruh endpoint daftar.
const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params adalah hasil normalisasi parameter paginasi dari query string.
type Params struct {
	Page    int
	PerPage int
}

// Parse menormalkan page dan per_page dari query string. Nilai per_page negatif
// diartikan sebagai permintaan tanpa paginasi (misalnya untuk ekspor).
func Parse(page, perPage string) Params {
	return Params{
		Page:    normalizePage(page),
		PerPage: normalizePerPage(perPage),
	}
}

// Offset menghitung offset baris untuk query database.
func (p Params) Offset() int {
	if p.Page <= 1 || p.PerPage <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

// TotalPages menghitung jumlah halaman dari total data.
func (p Params) TotalPages(total int) int {
	if p.PerPage <= 0 {
		return 1
	}
	pages := total / p.PerPage
	if total%p.PerPage != 0 {
		pages++
	}
	if pages == 0 {
		return 1
	}
	return pages
}

func normalizePage(raw string) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return DefaultPage
	}
	return value
}

func normalizePerPage(raw string) int {
	if raw == "" {
		return DefaultPerPage
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return DefaultPerPage
	}
	if value < 0 {
		return -1
	}
	if value == 0 {
		return DefaultPerPage
	}
	if value > MaxPerPage {
		return MaxPerPage
	}
	return value
}
