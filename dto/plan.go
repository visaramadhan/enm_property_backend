package dto

import (
	"time"

	"enmproperty/backend/models"
	"enmproperty/backend/pkg/money"
)

// PlanView adalah representasi paket langganan untuk halaman pricing dan admin.
type PlanView struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
	PriceLabel   string    `json:"price_label"`
	DurationDays int       `json:"duration_days"`
	ListingQuota int       `json:"listing_quota"`
	IsUnlimited  bool      `json:"is_unlimited"`
	Features     []string  `json:"features"`
	IsFree       bool      `json:"is_free"`
	IsActive     bool      `json:"is_active"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewPlanView membentuk tampilan paket dari entitas domain.
func NewPlanView(plan *models.Plan) PlanView {
	if plan == nil {
		return PlanView{}
	}
	features := plan.Features
	if features == nil {
		features = []string{}
	}
	return PlanView{
		ID:           plan.ID,
		Slug:         plan.Slug,
		Name:         plan.Name,
		Description:  plan.Description,
		Price:        plan.Price,
		PriceLabel:   money.Rupiah(plan.Price),
		DurationDays: plan.DurationDays,
		ListingQuota: plan.ListingQuota,
		IsUnlimited:  plan.IsUnlimited(),
		Features:     features,
		IsFree:       plan.IsFree,
		IsActive:     plan.IsActive,
		SortOrder:    plan.SortOrder,
		CreatedAt:    plan.CreatedAt,
		UpdatedAt:    plan.UpdatedAt,
	}
}

// NewPlanViews membentuk daftar tampilan paket.
func NewPlanViews(plans []*models.Plan) []PlanView {
	views := make([]PlanView, 0, len(plans))
	for _, plan := range plans {
		views = append(views, NewPlanView(plan))
	}
	return views
}

// CreatePlanRequest adalah body POST /admin/packages.
type CreatePlanRequest struct {
	Slug         string   `json:"slug" validate:"omitempty,max=60"`
	Name         string   `json:"name" validate:"required,min=3,max=100"`
	Description  string   `json:"description" validate:"omitempty,max=500"`
	Price        int64    `json:"price" validate:"gte=0"`
	DurationDays int      `json:"duration_days" validate:"required,gte=1,lte=3650"`
	ListingQuota int      `json:"listing_quota" validate:"required,gte=-1"`
	Features     []string `json:"features" validate:"omitempty,dive,max=120"`
	IsFree       bool     `json:"is_free"`
	IsActive     bool     `json:"is_active"`
	SortOrder    int      `json:"sort_order" validate:"gte=0"`
}

// UpdatePlanRequest adalah body PUT /admin/packages/{id}.
type UpdatePlanRequest struct {
	Name         string   `json:"name" validate:"omitempty,min=3,max=100"`
	Description  *string  `json:"description" validate:"omitempty,max=500"`
	Price        *int64   `json:"price" validate:"omitempty,gte=0"`
	DurationDays *int     `json:"duration_days" validate:"omitempty,gte=1,lte=3650"`
	ListingQuota *int     `json:"listing_quota" validate:"omitempty,gte=-1"`
	Features     []string `json:"features" validate:"omitempty,dive,max=120"`
	IsFree       *bool    `json:"is_free"`
	IsActive     *bool    `json:"is_active"`
	SortOrder    *int     `json:"sort_order" validate:"omitempty,gte=0"`
}

// SetPlanStatusRequest adalah body PATCH /admin/packages/{id}/status.
type SetPlanStatusRequest struct {
	Active bool `json:"active"`
}

// ReorderPlansRequest adalah body PATCH /admin/packages/order.
type ReorderPlansRequest struct {
	PlanIDs []string `json:"plan_ids" validate:"required,min=1,dive,required"`
}
