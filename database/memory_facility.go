package database

import (
	"errors"
	"strings"

	"enmproperty/backend/models"
)

type memoryFacilityRepository struct{ store *Store }

func (r *memoryFacilityRepository) Create(facility *models.Facility) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, existing := range r.store.facilities {
		if equalFold(existing.Slug, facility.Slug) {
			return errors.New("slug fasilitas sudah dipakai")
		}
	}
	r.store.facilities[facility.ID] = copyFacility(facility)
	return nil
}

func (r *memoryFacilityRepository) Update(facility *models.Facility) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.facilities[facility.ID]; !exists {
		return ErrNotFound
	}
	r.store.facilities[facility.ID] = copyFacility(facility)
	return nil
}

func (r *memoryFacilityRepository) Delete(id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.facilities[id]; !exists {
		return ErrNotFound
	}
	delete(r.store.facilities, id)
	return nil
}

func (r *memoryFacilityRepository) GetByID(id string) (*models.Facility, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	facility, exists := r.store.facilities[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyFacility(facility), nil
}

func (r *memoryFacilityRepository) GetBySlug(slug string) (*models.Facility, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, facility := range r.store.facilities {
		if equalFold(facility.Slug, slug) {
			return copyFacility(facility), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryFacilityRepository) SlugExists(slug, excludeID string) bool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, facility := range r.store.facilities {
		if facility.ID == excludeID {
			continue
		}
		if equalFold(facility.Slug, slug) {
			return true
		}
	}
	return false
}

func (r *memoryFacilityRepository) List(q ListQuery) ([]*models.Facility, int, error) {
	items := make([]*models.Facility, 0)
	r.store.mu.RLock()
	for _, facility := range r.store.facilities {
		items = append(items, copyFacility(facility))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Facility, 0, len(items))
	for _, facility := range items {
		if q.Filter("is_active") == "true" && !facility.IsActive {
			continue
		}
		if q.Filter("is_active") == "false" && facility.IsActive {
			continue
		}
		if q.Search != "" && !ContainsFold(facility.Name, q.Search) &&
			!ContainsFold(facility.Description, q.Search) {
			continue
		}
		filtered = append(filtered, facility)
	}

	SortByFunc(filtered, func(a, b *models.Facility) bool {
		if a.SortOrder != b.SortOrder {
			return a.SortOrder < b.SortOrder
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}, false)

	page, total := Paginate(filtered, q)
	return page, total, nil
}
