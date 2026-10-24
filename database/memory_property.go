package database

import (
	"sort"
	"strconv"
	"time"

	"enmproperty/backend/models"
)

type memoryPropertyRepository struct{ store *Store }

// attachPhotos melengkapi properti dengan daftar foto dari koleksi photos.
// Pemanggil wajib sudah memegang lock Store.
func (r *memoryPropertyRepository) attachPhotos(property *models.Property) *models.Property {
	if property == nil {
		return nil
	}
	photos := make([]models.PropertyPhoto, 0, len(r.store.photos))
	for _, photo := range r.store.photos {
		if photo.PropertyID == property.ID {
			photos = append(photos, *copyPhoto(photo))
		}
	}
	sort.SliceStable(photos, func(i, j int) bool {
		if photos[i].SortOrder != photos[j].SortOrder {
			return photos[i].SortOrder < photos[j].SortOrder
		}
		return photos[i].CreatedAt.Before(photos[j].CreatedAt)
	})
	property.Photos = photos
	return property
}

func (r *memoryPropertyRepository) Create(property *models.Property) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.properties[property.ID] = copyProperty(property)
	for i := range property.Photos {
		r.storePhotoLocked(property.ID, &property.Photos[i])
	}
	return nil
}

func (r *memoryPropertyRepository) Update(property *models.Property) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.properties[property.ID]; !exists {
		return ErrNotFound
	}
	r.store.properties[property.ID] = copyProperty(property)
	for i := range property.Photos {
		r.storePhotoLocked(property.ID, &property.Photos[i])
	}
	return nil
}

// storePhotoLocked menyimpan foto dengan PropertyID yang sudah dipastikan.
func (r *memoryPropertyRepository) storePhotoLocked(propertyID string, photo *models.PropertyPhoto) {
	stored := *photo
	if stored.PropertyID == "" {
		stored.PropertyID = propertyID
	}
	r.store.photos[stored.ID] = &stored
}

// Delete melakukan penghapusan lunak agar riwayat laporan tetap utuh.
func (r *memoryPropertyRepository) Delete(id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	property, exists := r.store.properties[id]
	if !exists {
		return ErrNotFound
	}
	now := time.Now()
	property.DeletedAt = &now
	property.UpdatedAt = now
	return nil
}

func (r *memoryPropertyRepository) GetByID(id string) (*models.Property, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	property, exists := r.store.properties[id]
	if !exists {
		return nil, ErrNotFound
	}
	return r.attachPhotos(copyProperty(property)), nil
}

func (r *memoryPropertyRepository) GetBySlug(slug string) (*models.Property, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, property := range r.store.properties {
		if equalFold(property.Slug, slug) {
			return r.attachPhotos(copyProperty(property)), nil
		}
	}
	return nil, ErrNotFound
}

func (r *memoryPropertyRepository) SlugExists(slug, excludeID string) bool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, property := range r.store.properties {
		if property.ID == excludeID {
			continue
		}
		if equalFold(property.Slug, slug) {
			return true
		}
	}
	return false
}

// CountByPartner menghitung properti mitra tanpa menyertakan yang terhapus.
// Tanpa status berarti menghitung seluruh properti mitra.
func (r *memoryPropertyRepository) CountByPartner(partnerID string, statuses ...models.PropertyStatus) int {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	count := 0
	for _, property := range r.store.properties {
		if property.PartnerID != partnerID || property.IsDeleted() {
			continue
		}
		if len(statuses) == 0 {
			count++
			continue
		}
		for _, status := range statuses {
			if property.Status == status {
				count++
				break
			}
		}
	}
	return count
}

// UpdateStatusByPartner memindahkan seluruh properti mitra dari satu status ke
// status lain, dipakai job penandaan kedaluwarsa langganan.
func (r *memoryPropertyRepository) UpdateStatusByPartner(partnerID string, from, to models.PropertyStatus) int {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	now := time.Now()
	updated := 0
	for _, property := range r.store.properties {
		if property.PartnerID != partnerID || property.IsDeleted() || property.Status != from {
			continue
		}
		property.Status = to
		property.UpdatedAt = now
		updated++
	}
	return updated
}

func (r *memoryPropertyRepository) List(q ListQuery) ([]*models.Property, int, error) {
	items := make([]*models.Property, 0)
	r.store.mu.RLock()
	for _, property := range r.store.properties {
		items = append(items, r.attachPhotos(copyProperty(property)))
	}
	r.store.mu.RUnlock()

	filtered := make([]*models.Property, 0, len(items))
	for _, property := range items {
		if property.IsDeleted() && q.Filter("include_deleted") != "true" {
			continue
		}
		if q.Filter("public") == "true" && !property.IsPublicVisible() {
			continue
		}
		if status := q.Filter("status"); status != "" && string(property.Status) != status {
			continue
		}
		if partnerID := q.Filter("partner_id"); partnerID != "" && property.PartnerID != partnerID {
			continue
		}
		if city := q.Filter("city"); city != "" && !equalFold(property.City, city) {
			continue
		}
		if province := q.Filter("province"); province != "" && !equalFold(property.Province, province) {
			continue
		}
		if listingType := q.Filter("listing_type"); listingType != "" && string(property.ListingType) != listingType {
			continue
		}
		if propertyType := q.Filter("property_type"); propertyType != "" && string(property.PropertyType) != propertyType {
			continue
		}
		if min := q.Filter("price_min"); min != "" {
			if value, err := strconv.ParseInt(min, 10, 64); err == nil && property.Price < value {
				continue
			}
		}
		if max := q.Filter("price_max"); max != "" {
			if value, err := strconv.ParseInt(max, 10, 64); err == nil && property.Price > value {
				continue
			}
		}
		if q.Search != "" && !ContainsFold(property.Title, q.Search) &&
			!ContainsFold(property.City, q.Search) && !ContainsFold(property.Address, q.Search) &&
			!ContainsFold(property.PartnerName, q.Search) {
			continue
		}
		filtered = append(filtered, property)
	}

	sortField := q.SortBy
	desc := q.SortDesc
	if sortField == "" {
		sortField = "created_at"
		desc = true
	}
	switch sortField {
	case "price":
		SortByFunc(filtered, func(a, b *models.Property) bool { return a.Price < b.Price }, desc)
	case "views":
		SortByFunc(filtered, func(a, b *models.Property) bool { return a.Views < b.Views }, desc)
	case "title":
		SortByFunc(filtered, func(a, b *models.Property) bool { return a.Title < b.Title }, desc)
	default:
		SortByFunc(filtered, func(a, b *models.Property) bool { return a.CreatedAt.Before(b.CreatedAt) }, desc)
	}

	page, total := Paginate(filtered, q)
	return page, total, nil
}

func (r *memoryPropertyRepository) CreatePhoto(photo *models.PropertyPhoto) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.storePhotoLocked(photo.PropertyID, photo)
	return nil
}

func (r *memoryPropertyRepository) UpdatePhoto(photo *models.PropertyPhoto) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.photos[photo.ID]; !exists {
		return ErrNotFound
	}
	r.store.photos[photo.ID] = copyPhoto(photo)
	return nil
}

func (r *memoryPropertyRepository) DeletePhoto(id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.photos[id]; !exists {
		return ErrNotFound
	}
	delete(r.store.photos, id)
	return nil
}

func (r *memoryPropertyRepository) GetPhoto(id string) (*models.PropertyPhoto, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	photo, exists := r.store.photos[id]
	if !exists {
		return nil, ErrNotFound
	}
	return copyPhoto(photo), nil
}

func (r *memoryPropertyRepository) ListPhotos(propertyID string) ([]*models.PropertyPhoto, error) {
	photos := make([]*models.PropertyPhoto, 0)
	r.store.mu.RLock()
	for _, photo := range r.store.photos {
		if photo.PropertyID == propertyID {
			photos = append(photos, copyPhoto(photo))
		}
	}
	r.store.mu.RUnlock()

	sort.SliceStable(photos, func(i, j int) bool {
		if photos[i].SortOrder != photos[j].SortOrder {
			return photos[i].SortOrder < photos[j].SortOrder
		}
		return photos[i].CreatedAt.Before(photos[j].CreatedAt)
	})
	return photos, nil
}
