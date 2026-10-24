package dto

import (
	"time"

	"enmproperty/backend/models"
)

// FavoriteView adalah satu baris favorit customer beserta ringkasan propertinya.
type FavoriteView struct {
	ID         string           `json:"id"`
	PropertyID string           `json:"property_id"`
	CreatedAt  time.Time        `json:"created_at"`
	Property   *PropertySummary `json:"property,omitempty"`
}

// NewFavoriteView membentuk tampilan favorit dari entitas domain.
func NewFavoriteView(favorite *models.Favorite, property *models.Property) FavoriteView {
	if favorite == nil {
		return FavoriteView{}
	}
	view := FavoriteView{ID: favorite.ID, PropertyID: favorite.PropertyID, CreatedAt: favorite.CreatedAt}
	if property != nil {
		summary := NewPropertySummary(property)
		summary.IsFavorite = true
		view.Property = &summary
	}
	return view
}

// FavoriteToggleResponse adalah respons POST dan DELETE /me/favorites/{propertyId}.
type FavoriteToggleResponse struct {
	PropertyID string `json:"property_id"`
	IsFavorite bool   `json:"is_favorite"`
}
