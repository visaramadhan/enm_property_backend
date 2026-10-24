// Package dto memuat struktur permintaan dan respons yang dipakai handler serta
// service agar kontrak API tetap konsisten dan terpisah dari entitas domain.
package dto

import (
	"strconv"

	"enmproperty/backend/database"
)

// PaginationQuery adalah parameter daftar standar dari query string.
type PaginationQuery struct {
	Page     int    `form:"page" json:"page"`
	PerPage  int    `form:"per_page" json:"per_page"`
	Search   string `form:"search" json:"search"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortDesc bool   `form:"sort_desc" json:"sort_desc"`
}

// ToListQuery mengubah parameter paginasi menjadi ListQuery repository.
func (q PaginationQuery) ToListQuery(filters map[string]string) database.ListQuery {
	return database.ListQuery{
		Search:   q.Search,
		Page:     q.Page,
		PerPage:  q.PerPage,
		SortBy:   q.SortBy,
		SortDesc: q.SortDesc,
		Filters:  filters,
	}
}

// ListResult adalah bentuk respons daftar dengan metadata paginasi.
type ListResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

// NewListResult membentuk respons daftar.
func NewListResult[T any](items []T, total int) ListResult[T] {
	if items == nil {
		items = []T{}
	}
	return ListResult[T]{Items: items, Total: total}
}

// IDParam adalah parameter path berisi ID.
type IDParam struct {
	ID string `uri:"id" validate:"required"`
}

// SlugParam adalah parameter path berisi slug.
type SlugParam struct {
	Slug string `uri:"slug" validate:"required"`
}

// CodeParam adalah parameter path berisi kode order.
type CodeParam struct {
	OrderCode string `uri:"orderCode" validate:"required"`
}

// PropertyIDParam adalah parameter path berisi ID properti.
type PropertyIDParam struct {
	PropertyID string `uri:"propertyId" validate:"required"`
}

// PhotoIDParam adalah parameter path berisi ID foto.
type PhotoIDParam struct {
	PropertyID string `uri:"propertyId" validate:"required"`
	PhotoID    string `uri:"photoId" validate:"required"`
}

// IntQuery mengambil nilai integer dari map filter yang sudah dinormalisasi.
func IntQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
