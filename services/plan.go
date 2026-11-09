package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/slug"
	"enmproperty/backend/platform/httpx"
)

// PlanService mengelola paket langganan beserta urutan dan status aktifnya (BE-05).
type PlanService struct {
	deps Deps
}

// ListActive mengembalikan paket aktif untuk halaman pricing.
func (s *PlanService) ListActive() ([]dto.PlanView, error) {
	plans, _, err := s.deps.Repos.Plans.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"is_active": "true"},
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return dto.NewPlanViews(plans), nil
}

// ListAll mengembalikan seluruh paket untuk halaman manajemen admin.
func (s *PlanService) ListAll(q dto.PaginationQuery) (dto.ListResult[dto.PlanView], error) {
	plans, total, err := s.deps.Repos.Plans.List(q.ToListQuery(nil))
	if err != nil {
		return dto.ListResult[dto.PlanView]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewPlanViews(plans), total), nil
}

// GetBySlug mengambil paket aktif berdasarkan slug untuk halaman checkout.
func (s *PlanService) GetBySlug(planSlug string) (dto.PlanView, error) {
	plan, err := s.deps.Repos.Plans.GetBySlug(strings.TrimSpace(planSlug))
	if err != nil {
		if database.IsNotFound(err) {
			return dto.PlanView{}, httpx.NotFound("Paket tidak ditemukan")
		}
		return dto.PlanView{}, httpx.Internal(err)
	}
	if !plan.IsActive {
		return dto.PlanView{}, httpx.NotFound("Paket sedang tidak tersedia")
	}
	return dto.NewPlanView(plan), nil
}

// GetByID mengambil entitas paket berdasarkan ID.
func (s *PlanService) GetByID(id string) (*models.Plan, error) {
	plan, err := s.deps.Repos.Plans.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Paket tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return plan, nil
}

// Create menambahkan paket baru.
func (s *PlanService) Create(req dto.CreatePlanRequest) (dto.PlanView, error) {
	planSlug := slug.Normalize(req.Slug)
	if planSlug == "" {
		planSlug = slug.Make(req.Name)
	}
	if planSlug == "" {
		return dto.PlanView{}, httpx.BadRequest("Slug paket tidak valid")
	}
	if _, err := s.deps.Repos.Plans.GetBySlug(planSlug); err == nil {
		return dto.PlanView{}, httpx.Conflict("Slug paket sudah dipakai")
	}

	now := timeNow()
	plan := &models.Plan{
		ID:             newID(),
		Slug:           planSlug,
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		Price:          req.Price,
		DurationDays:   req.DurationDays,
		ListingQuota:   req.ListingQuota,
		Features:       cleanList(req.Features),
		IsFree:         req.IsFree || req.Price == 0,
		IsActive:       req.IsActive,
		SortOrder:      req.SortOrder,
		Level:          req.Level,
		BoostEnabled:   req.BoostEnabled,
		BoostQuota:     req.BoostQuota,
		KeywordAccess:  req.KeywordAccess,
		CatalogAccess:  req.CatalogAccess,
		WhatsAppAccess: req.WhatsAppAccess,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	// Paket gratis tidak boleh memiliki akses lanjutan maupun boost.
	plan.ApplyAccessRules()
	if err := s.deps.Repos.Plans.Create(plan); err != nil {
		return dto.PlanView{}, httpx.Conflict("Paket tidak dapat disimpan, slug mungkin sudah dipakai")
	}
	return dto.NewPlanView(plan), nil
}

// Update mengubah sebagian data paket.
func (s *PlanService) Update(id string, req dto.UpdatePlanRequest) (dto.PlanView, error) {
	plan, err := s.GetByID(id)
	if err != nil {
		return dto.PlanView{}, err
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		plan.Name = name
	}
	if req.Description != nil {
		plan.Description = strings.TrimSpace(*req.Description)
	}
	if req.Price != nil {
		plan.Price = *req.Price
		if *req.Price == 0 {
			plan.IsFree = true
		}
	}
	if req.DurationDays != nil {
		plan.DurationDays = *req.DurationDays
	}
	if req.ListingQuota != nil {
		plan.ListingQuota = *req.ListingQuota
	}
	if req.Features != nil {
		plan.Features = cleanList(req.Features)
	}
	if req.IsFree != nil {
		plan.IsFree = *req.IsFree
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.Level != nil {
		plan.Level = *req.Level
	}
	if req.BoostEnabled != nil {
		plan.BoostEnabled = *req.BoostEnabled
	}
	if req.BoostQuota != nil {
		plan.BoostQuota = *req.BoostQuota
	}
	if req.KeywordAccess != nil {
		plan.KeywordAccess = *req.KeywordAccess
	}
	if req.CatalogAccess != nil {
		plan.CatalogAccess = *req.CatalogAccess
	}
	if req.WhatsAppAccess != nil {
		plan.WhatsAppAccess = *req.WhatsAppAccess
	}
	// Aturan paket gratis ditegakkan terakhir agar tidak dapat dilanggar API.
	plan.ApplyAccessRules()
	plan.UpdatedAt = timeNow()

	if err := s.deps.Repos.Plans.Update(plan); err != nil {
		return dto.PlanView{}, httpx.Internal(err)
	}
	return dto.NewPlanView(plan), nil
}

// SetActive mengaktifkan atau menonaktifkan paket.
func (s *PlanService) SetActive(id string, active bool) (dto.PlanView, error) {
	plan, err := s.GetByID(id)
	if err != nil {
		return dto.PlanView{}, err
	}
	plan.IsActive = active
	plan.UpdatedAt = timeNow()
	if err := s.deps.Repos.Plans.Update(plan); err != nil {
		return dto.PlanView{}, httpx.Internal(err)
	}
	return dto.NewPlanView(plan), nil
}

// Reorder menetapkan urutan paket sesuai daftar ID yang dikirim.
func (s *PlanService) Reorder(req dto.ReorderPlansRequest) ([]dto.PlanView, error) {
	for index, id := range req.PlanIDs {
		plan, err := s.deps.Repos.Plans.GetByID(id)
		if err != nil {
			if database.IsNotFound(err) {
				return nil, httpx.NotFound("Paket pada daftar urutan tidak ditemukan")
			}
			return nil, httpx.Internal(err)
		}
		plan.SortOrder = index + 1
		plan.UpdatedAt = timeNow()
		if err := s.deps.Repos.Plans.Update(plan); err != nil {
			return nil, httpx.Internal(err)
		}
	}
	return s.ListActive()
}

// cleanList memangkas spasi dan membuang entri kosong pada daftar teks.
func cleanList(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 {
		return []string{}
	}
	return cleaned
}
