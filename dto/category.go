package dto

import (
	"time"

	"enmproperty/backend/models"
)

// CategoryView adalah representasi kategori properti untuk katalog dan master data.
type CategoryView struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewCategoryView membentuk tampilan kategori dari entitas domain.
func NewCategoryView(category *models.Category) CategoryView {
	if category == nil {
		return CategoryView{}
	}
	return CategoryView{
		ID:          category.ID,
		Slug:        category.Slug,
		Name:        category.Name,
		Description: category.Description,
		SortOrder:   category.SortOrder,
		IsActive:    category.IsActive,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

// NewCategoryViews membentuk daftar tampilan kategori.
func NewCategoryViews(categories []*models.Category) []CategoryView {
	views := make([]CategoryView, 0, len(categories))
	for _, category := range categories {
		views = append(views, NewCategoryView(category))
	}
	return views
}

// CreateCategoryRequest adalah body POST /admin/categories.
type CreateCategoryRequest struct {
	Slug        string `json:"slug" validate:"omitempty,max=60"`
	Name        string `json:"name" validate:"required,min=2,max=80"`
	Description string `json:"description" validate:"omitempty,max=300"`
	SortOrder   int    `json:"sort_order" validate:"gte=0"`
	IsActive    bool   `json:"is_active"`
}

// UpdateCategoryRequest adalah body PUT /admin/categories/{id}.
type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=80"`
	Description *string `json:"description" validate:"omitempty,max=300"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}
