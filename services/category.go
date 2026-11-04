package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/slug"
	"enmproperty/backend/platform/httpx"
)

// CategoryService mengelola kategori properti untuk katalog publik dan master
// data superadmin.
type CategoryService struct {
	deps Deps
}

// ListActive mengembalikan kategori aktif untuk filter katalog publik.
func (s *CategoryService) ListActive() ([]dto.CategoryView, error) {
	categories, _, err := s.deps.Repos.Categories.List(database.ListQuery{
		PerPage: 0,
		Filters: map[string]string{"is_active": "true"},
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return dto.NewCategoryViews(categories), nil
}

// ListAll mengembalikan seluruh kategori untuk halaman master data.
func (s *CategoryService) ListAll(q dto.PaginationQuery) (dto.ListResult[dto.CategoryView], error) {
	categories, total, err := s.deps.Repos.Categories.List(q.ToListQuery(nil))
	if err != nil {
		return dto.ListResult[dto.CategoryView]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewCategoryViews(categories), total), nil
}

// GetByID mengambil entitas kategori berdasarkan ID.
func (s *CategoryService) GetByID(id string) (*models.Category, error) {
	category, err := s.deps.Repos.Categories.GetByID(strings.TrimSpace(id))
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Kategori tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return category, nil
}

// Create menambahkan kategori baru.
func (s *CategoryService) Create(req dto.CreateCategoryRequest) (dto.CategoryView, error) {
	categorySlug := slug.Normalize(req.Slug)
	if categorySlug == "" {
		categorySlug = slug.Make(req.Name)
	}
	if categorySlug == "" {
		return dto.CategoryView{}, httpx.BadRequest("Slug kategori tidak valid")
	}
	if s.deps.Repos.Categories.SlugExists(categorySlug, "") {
		return dto.CategoryView{}, httpx.Conflict("Slug kategori sudah dipakai")
	}

	now := timeNow()
	category := &models.Category{
		ID:          newID(),
		Slug:        categorySlug,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.deps.Repos.Categories.Create(category); err != nil {
		return dto.CategoryView{}, httpx.Conflict("Kategori tidak dapat disimpan, slug mungkin sudah dipakai")
	}
	return dto.NewCategoryView(category), nil
}

// Update mengubah sebagian data kategori.
func (s *CategoryService) Update(id string, req dto.UpdateCategoryRequest) (dto.CategoryView, error) {
	category, err := s.GetByID(id)
	if err != nil {
		return dto.CategoryView{}, err
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		category.Name = name
	}
	if req.Description != nil {
		category.Description = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	category.UpdatedAt = timeNow()

	if err := s.deps.Repos.Categories.Update(category); err != nil {
		return dto.CategoryView{}, httpx.Internal(err)
	}
	return dto.NewCategoryView(category), nil
}

// SetActive mengaktifkan atau menonaktifkan kategori.
func (s *CategoryService) SetActive(id string, active bool) (dto.CategoryView, error) {
	category, err := s.GetByID(id)
	if err != nil {
		return dto.CategoryView{}, err
	}
	category.IsActive = active
	category.UpdatedAt = timeNow()
	if err := s.deps.Repos.Categories.Update(category); err != nil {
		return dto.CategoryView{}, httpx.Internal(err)
	}
	return dto.NewCategoryView(category), nil
}

// Delete menghapus kategori dari master data.
func (s *CategoryService) Delete(id string) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	if err := s.deps.Repos.Categories.Delete(id); err != nil {
		if database.IsNotFound(err) {
			return httpx.NotFound("Kategori tidak ditemukan")
		}
		return httpx.Internal(err)
	}
	return nil
}
