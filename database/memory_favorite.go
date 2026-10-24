package database

import (
	"enmproperty/backend/models"
)

type memoryFavoriteRepository struct{ store *Store }

// Add menyimpan favorit dan mengabaikan duplikasi user-properti.
func (r *memoryFavoriteRepository) Add(favorite *models.Favorite) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, existing := range r.store.favorites {
		if existing.UserID == favorite.UserID && existing.PropertyID == favorite.PropertyID {
			return nil
		}
	}
	r.store.favorites[favorite.ID] = copyFavorite(favorite)
	return nil
}

func (r *memoryFavoriteRepository) Remove(userID, propertyID string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for id, favorite := range r.store.favorites {
		if favorite.UserID == userID && favorite.PropertyID == propertyID {
			delete(r.store.favorites, id)
			return nil
		}
	}
	return ErrNotFound
}

func (r *memoryFavoriteRepository) Exists(userID, propertyID string) bool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, favorite := range r.store.favorites {
		if favorite.UserID == userID && favorite.PropertyID == propertyID {
			return true
		}
	}
	return false
}

func (r *memoryFavoriteRepository) ListByUser(userID string) ([]*models.Favorite, error) {
	items := make([]*models.Favorite, 0)
	r.store.mu.RLock()
	for _, favorite := range r.store.favorites {
		if favorite.UserID == userID {
			items = append(items, copyFavorite(favorite))
		}
	}
	r.store.mu.RUnlock()

	SortByFunc(items, func(a, b *models.Favorite) bool {
		return a.CreatedAt.Before(b.CreatedAt)
	}, true)
	return items, nil
}

func (r *memoryFavoriteRepository) CountByProperty(propertyID string) int {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	count := 0
	for _, favorite := range r.store.favorites {
		if favorite.PropertyID == propertyID {
			count++
		}
	}
	return count
}
