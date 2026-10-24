package models

import "time"

// Property adalah listing properti milik mitra.
type Property struct {
	ID             string          `json:"id"`
	PartnerID      string          `json:"partner_id"`
	PartnerName    string          `json:"partner_name"`
	Slug           string          `json:"slug"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	ListingType    ListingType     `json:"listing_type"`
	PropertyType   PropertyType    `json:"property_type"`
	Address        string          `json:"address"`
	City           string          `json:"city"`
	Province       string          `json:"province"`
	District       string          `json:"district"`
	Latitude       float64         `json:"latitude"`
	Longitude      float64         `json:"longitude"`
	Price          int64           `json:"price"`
	PricePeriod    PricePeriod     `json:"price_period,omitempty"`
	LandArea       int             `json:"land_area"`
	BuildingArea   int             `json:"building_area"`
	Bedrooms       int             `json:"bedrooms"`
	Bathrooms      int             `json:"bathrooms"`
	Facilities     []string        `json:"facilities"`
	UnitsTotal     int             `json:"units_total"`
	UnitsAvailable int             `json:"units_available"`
	Status         PropertyStatus  `json:"status"`
	Views          int             `json:"views"`
	ContactCount   int             `json:"contact_count"`
	Photos         []PropertyPhoto `json:"photos,omitempty"`
	PublishedAt    *time.Time      `json:"published_at,omitempty"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// IsDeleted memeriksa apakah properti sudah dihapus secara lunak.
func (p *Property) IsDeleted() bool { return p.DeletedAt != nil }

// IsPublicVisible memeriksa apakah properti tampil di katalog publik.
func (p *Property) IsPublicVisible() bool { return p.Status == PropertyStatusActive }

// PrimaryPhoto mengembalikan foto utama, atau foto pertama bila belum ditandai.
func (p *Property) PrimaryPhoto() *PropertyPhoto {
	if len(p.Photos) == 0 {
		return nil
	}
	for i := range p.Photos {
		if p.Photos[i].IsPrimary {
			return &p.Photos[i]
		}
	}
	return &p.Photos[0]
}

// CoverURL mengembalikan URL foto utama untuk kartu katalog.
func (p *Property) CoverURL() string {
	if photo := p.PrimaryPhoto(); photo != nil {
		return photo.URL
	}
	return ""
}

// HasUnits memeriksa apakah unit masih tersedia untuk dihubungi.
func (p *Property) HasUnits() bool { return p.UnitsAvailable != 0 }

// PropertyPhoto adalah foto properti yang disimpan pada storage.
type PropertyPhoto struct {
	ID          string    `json:"id"`
	PropertyID  string    `json:"property_id"`
	URL         string    `json:"url"`
	StoragePath string    `json:"storage_path"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	SortOrder   int       `json:"sort_order"`
	IsPrimary   bool      `json:"is_primary"`
	CreatedAt   time.Time `json:"created_at"`
}

// Favorite adalah properti yang disimpan customer.
type Favorite struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	PropertyID string    `json:"property_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// PropertyView mencatat satu tayangan halaman detail properti untuk statistik.
type PropertyView struct {
	ID         string    `json:"id"`
	PropertyID string    `json:"property_id"`
	PartnerID  string    `json:"partner_id"`
	ViewedAt   time.Time `json:"viewed_at"`
}

// PropertyContact mencatat satu lead WhatsApp yang diteruskan ke mitra.
type PropertyContact struct {
	ID         string    `json:"id"`
	PropertyID string    `json:"property_id"`
	PartnerID  string    `json:"partner_id"`
	UserID     string    `json:"user_id"`
	Channel    string    `json:"channel"`
	CreatedAt  time.Time `json:"created_at"`
}
