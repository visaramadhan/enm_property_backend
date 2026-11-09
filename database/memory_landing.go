package database

import (
	"sort"

	"enmproperty/backend/models"
)

type memoryLandingRepository struct{ store *Store }

func (r *memoryLandingRepository) All() []*models.LandingContent {
	r.store.mu.RLock()
	items := make([]*models.LandingContent, 0, len(r.store.landing))
	for _, item := range r.store.landing {
		items = append(items, copyLanding(item))
	}
	r.store.mu.RUnlock()

	sort.SliceStable(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	return items
}

func (r *memoryLandingRepository) Get(key string) (*models.LandingContent, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	item, exists := r.store.landing[key]
	if !exists {
		return nil, ErrNotFound
	}
	return copyLanding(item), nil
}

func (r *memoryLandingRepository) Set(content *models.LandingContent) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.landing[content.Key] = copyLanding(content)
	return nil
}
