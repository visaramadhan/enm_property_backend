package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/slug"
	"enmproperty/backend/platform/httpx"
)

// FacilityService mengelola master data fasilitas properti untuk pilihan mitra
// dan halaman master data superadmin.
type FacilityService struct {
	deps Deps
}

// ListActive mengembalikan fasilitas aktif untuk pilihan mitra saat mengunggah
// listing.
func (s *FacilityService) ListActive() ([]dto.FacilityView, error) {
	facilities, _, err := s.deps.Repos.Facilities.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"is_active": "true"},
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return dto.NewFacilityViews(facilities), nil
}

// ListAll mengembalikan seluruh fasilitas untuk halaman master data.
func (s *FacilityService) ListAll(q dto.PaginationQuery) (dto.ListResult[dto.FacilityView], error) {
	facilities, total, err := s.deps.Repos.Facilities.List(q.ToListQuery(nil))
	if err != nil {
		return dto.ListResult[dto.FacilityView]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewFacilityViews(facilities), total), nil
}

// GetByID mengambil entitas fasilitas berdasarkan ID.
func (s *FacilityService) GetByID(id string) (*models.Facility, error) {
	facility, err := s.deps.Repos.Facilities.GetByID(strings.TrimSpace(id))
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Fasilitas tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return facility, nil
}

// Create menambahkan fasilitas baru.
func (s *FacilityService) Create(req dto.CreateFacilityRequest) (dto.FacilityView, error) {
	facilitySlug := slug.Normalize(req.Slug)
	if facilitySlug == "" {
		facilitySlug = slug.Make(req.Name)
	}
	if facilitySlug == "" {
		return dto.FacilityView{}, httpx.BadRequest("Slug fasilitas tidak valid")
	}
	if s.deps.Repos.Facilities.SlugExists(facilitySlug, "") {
		return dto.FacilityView{}, httpx.Conflict("Slug fasilitas sudah dipakai")
	}

	now := timeNow()
	facility := &models.Facility{
		ID:          newID(),
		Slug:        facilitySlug,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Icon:        strings.TrimSpace(req.Icon),
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.deps.Repos.Facilities.Create(facility); err != nil {
		return dto.FacilityView{}, httpx.Conflict("Fasilitas tidak dapat disimpan, slug mungkin sudah dipakai")
	}
	return dto.NewFacilityView(facility), nil
}

// Update mengubah sebagian data fasilitas.
func (s *FacilityService) Update(id string, req dto.UpdateFacilityRequest) (dto.FacilityView, error) {
	facility, err := s.GetByID(id)
	if err != nil {
		return dto.FacilityView{}, err
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		facility.Name = name
	}
	if req.Description != nil {
		facility.Description = strings.TrimSpace(*req.Description)
	}
	if req.Icon != nil {
		facility.Icon = strings.TrimSpace(*req.Icon)
	}
	if req.SortOrder != nil {
		facility.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		facility.IsActive = *req.IsActive
	}
	facility.UpdatedAt = timeNow()

	if err := s.deps.Repos.Facilities.Update(facility); err != nil {
		return dto.FacilityView{}, httpx.Internal(err)
	}
	return dto.NewFacilityView(facility), nil
}

// SetActive mengaktifkan atau menonaktifkan fasilitas.
func (s *FacilityService) SetActive(id string, active bool) (dto.FacilityView, error) {
	facility, err := s.GetByID(id)
	if err != nil {
		return dto.FacilityView{}, err
	}
	facility.IsActive = active
	facility.UpdatedAt = timeNow()
	if err := s.deps.Repos.Facilities.Update(facility); err != nil {
		return dto.FacilityView{}, httpx.Internal(err)
	}
	return dto.NewFacilityView(facility), nil
}

// Delete menghapus fasilitas dari master data.
func (s *FacilityService) Delete(id string) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	if err := s.deps.Repos.Facilities.Delete(id); err != nil {
		if database.IsNotFound(err) {
			return httpx.NotFound("Fasilitas tidak ditemukan")
		}
		return httpx.Internal(err)
	}
	return nil
}
