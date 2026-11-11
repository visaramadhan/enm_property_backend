package dto

import (
	"time"

	"enmproperty/backend/models"
)

// FacilityView adalah representasi fasilitas properti untuk pilihan mitra dan
// master data superadmin.
type FacilityView struct {
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

// NewFacilityView membentuk tampilan fasilitas dari entitas domain.
func NewFacilityView(facility *models.Facility) FacilityView {
	if facility == nil {
		return FacilityView{}
	}
	return FacilityView{
		ID:          facility.ID,
		Slug:        facility.Slug,
		Name:        facility.Name,
		Description: facility.Description,
		Icon:        facility.Icon,
		SortOrder:   facility.SortOrder,
		IsActive:    facility.IsActive,
		CreatedAt:   facility.CreatedAt,
		UpdatedAt:   facility.UpdatedAt,
	}
}

// NewFacilityViews membentuk daftar tampilan fasilitas.
func NewFacilityViews(facilities []*models.Facility) []FacilityView {
	views := make([]FacilityView, 0, len(facilities))
	for _, facility := range facilities {
		views = append(views, NewFacilityView(facility))
	}
	return views
}

// CreateFacilityRequest adalah body POST /admin/facilities.
type CreateFacilityRequest struct {
	Slug        string `json:"slug" validate:"omitempty,max=60"`
	Name        string `json:"name" validate:"required,min=2,max=80"`
	Description string `json:"description" validate:"omitempty,max=300"`
	Icon        string `json:"icon" validate:"omitempty,max=60"`
	SortOrder   int    `json:"sort_order" validate:"gte=0"`
	IsActive    bool   `json:"is_active"`
}

// UpdateFacilityRequest adalah body PUT /admin/facilities/{id}.
type UpdateFacilityRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2,max=80"`
	Description *string `json:"description" validate:"omitempty,max=300"`
	Icon        *string `json:"icon" validate:"omitempty,max=60"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}
