package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// FavoriteService mengelola daftar properti favorit milik customer (BE-12).
type FavoriteService struct {
	deps Deps
}

// List mengembalikan properti favorit customer beserta ringkasan propertinya.
func (s *FavoriteService) List(user *models.User) (dto.ListResult[dto.FavoriteView], error) {
	if user == nil {
		return dto.ListResult[dto.FavoriteView]{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	favorites, err := s.deps.Repos.Favorites.ListByUser(user.ID)
	if err != nil {
		return dto.ListResult[dto.FavoriteView]{}, httpx.Internal(err)
	}

	views := make([]dto.FavoriteView, 0, len(favorites))
	for _, favorite := range favorites {
		property, err := s.deps.Repos.Properties.GetByID(favorite.PropertyID)
		if err != nil {
			// Properti yang sudah dihapus tidak ditampilkan pada daftar favorit.
			if database.IsNotFound(err) {
				continue
			}
			return dto.ListResult[dto.FavoriteView]{}, httpx.Internal(err)
		}
		if property.IsDeleted() {
			continue
		}
		views = append(views, dto.NewFavoriteView(favorite, property))
	}
	return dto.NewListResult(views, len(views)), nil
}

// Add menyimpan properti ke daftar favorit. Idempoten.
func (s *FavoriteService) Add(user *models.User, propertyID string) (dto.FavoriteToggleResponse, error) {
	if user == nil {
		return dto.FavoriteToggleResponse{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	property, err := s.deps.Repos.Properties.GetByID(strings.TrimSpace(propertyID))
	if err != nil {
		if database.IsNotFound(err) {
			return dto.FavoriteToggleResponse{}, httpx.NotFound("Properti tidak ditemukan")
		}
		return dto.FavoriteToggleResponse{}, httpx.Internal(err)
	}
	if property.IsDeleted() {
		return dto.FavoriteToggleResponse{}, httpx.NotFound("Properti tidak ditemukan")
	}

	favorite := &models.Favorite{
		ID:         newID(),
		UserID:     user.ID,
		PropertyID: property.ID,
		CreatedAt:  timeNow(),
	}
	if err := s.deps.Repos.Favorites.Add(favorite); err != nil {
		return dto.FavoriteToggleResponse{}, httpx.Internal(err)
	}
	return dto.FavoriteToggleResponse{PropertyID: property.ID, IsFavorite: true}, nil
}

// Remove menghapus properti dari daftar favorit. Idempoten.
func (s *FavoriteService) Remove(user *models.User, propertyID string) (dto.FavoriteToggleResponse, error) {
	if user == nil {
		return dto.FavoriteToggleResponse{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	key := strings.TrimSpace(propertyID)
	if err := s.deps.Repos.Favorites.Remove(user.ID, key); err != nil && !database.IsNotFound(err) {
		return dto.FavoriteToggleResponse{}, httpx.Internal(err)
	}
	return dto.FavoriteToggleResponse{PropertyID: key, IsFavorite: false}, nil
}

// Exists memeriksa apakah properti sudah ada pada daftar favorit pengguna.
func (s *FavoriteService) Exists(userID, propertyID string) bool {
	if userID == "" || propertyID == "" {
		return false
	}
	return s.deps.Repos.Favorites.Exists(userID, propertyID)
}
