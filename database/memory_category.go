package database

import (
	"errors"
	"strings"

	"enmproperty/backend/models"
)

type memoryCategoryRepository struct{ store *Store }

func (r *memoryCategoryRepository) Create(category *models.Category) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, existing := range r.store.categories {
		if equalFold(existing.Slug, category.Slug) {
			return errors.New("slug kategori sudah dipakai")
		}
	}
	r.store.categories[category.ID] = copyCategory(category)
	return nil
}

func (r *memoryCategoryRepository) Update(category *models.Category) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.categories[category.ID]; !exists {
		return ErrNotFound
	}
	r.store.categories[category.ID] = copyCategory(category)
	return nil
}

// Delete menghapus kategori. Kategori masih dipertahankan sebagai data master
// sehingga penghapusan dilakukan permanen (tidak ada relasi wajib properti).
func (r *memoryCategoryRepository) Delete(id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.categories[id]; !exists {
		return ErrNotFound
	}
	delete(r.store.categories, id)
	return nil
}

func (r *memoryCategoryRepository) GetByID(id string) (*models.Category, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	category, exists := r.store.categories[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyCategory(category), nil
}

func (r *memoryCategoryRepository) GetBySlug(slug string) (*models.Category, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, category := range r.store.categories {
		if equalFold(category.Slug, slug) {
			return copyCategory(category), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryCategoryRepository) SlugExists(slug, excludeID string) bool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, category := range r.store.categories {
		if category.ID == excludeID {
			continue
		}
		if equalFold(category.Slug, slug) {
			return true
		}
	}
	return false
}

func (r *memoryCategoryRepository) List(q ListQuery) ([]*models.Category, int, error) {
	items := make([]*models.Category, 0)
	r.store.mu.RLock()
	for _, category := range r.store.categories {
		items = append(items, copyCategory(category))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Category, 0, len(items))
	for _, category := range items {
		if q.Filter("is_active") == "true" && !category.IsActive {
			continue
		}
		if q.Filter("is_active") == "false" && category.IsActive {
			continue
		}
		if q.Search != "" && !ContainsFold(category.Name, q.Search) &&
			!ContainsFold(category.Description, q.Search) {
			continue
		}
		filtered = append(filtered, category)
	}

	SortByFunc(filtered, func(a, b *models.Category) bool {
		if a.SortOrder != b.SortOrder {
			return a.SortOrder < b.SortOrder
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}, false)

	page, total := Paginate(filtered, q)
	return page, total, nil
}
