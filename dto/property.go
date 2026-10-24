package dto

import (
	"strconv"
	"time"

	"enmproperty/backend/database"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/money"
)

// PhotoView adalah representasi satu foto properti.
type PhotoView struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// NewPhotoView membentuk tampilan foto dari entitas domain.
func NewPhotoView(photo *models.PropertyPhoto) PhotoView {
	if photo == nil {
		return PhotoView{}
	}
	return PhotoView{
		ID:        photo.ID,
		URL:       photo.URL,
		MimeType:  photo.MimeType,
		SizeBytes: photo.SizeBytes,
		Width:     photo.Width,
		Height:    photo.Height,
		SortOrder: photo.SortOrder,
		IsPrimary: photo.IsPrimary,
		CreatedAt: photo.CreatedAt,
	}
}

// NewPhotoViews membentuk daftar tampilan foto.
func NewPhotoViews(photos []models.PropertyPhoto) []PhotoView {
	views := make([]PhotoView, 0, len(photos))
	for i := range photos {
		views = append(views, NewPhotoView(&photos[i]))
	}
	return views
}

// PartnerSummary adalah ringkasan mitra yang ditempelkan pada properti.
type PartnerSummary struct {
	ID           string `json:"id"`
	BusinessName string `json:"business_name"`
	PICName      string `json:"pic_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	City         string `json:"city"`
	Province     string `json:"province"`
}

// PropertySummary adalah bentuk ringkas properti untuk kartu katalog dan tabel.
type PropertySummary struct {
	ID                string     `json:"id"`
	Slug              string     `json:"slug"`
	Title             string     `json:"title"`
	ListingType       string     `json:"listing_type"`
	ListingTypeLabel  string     `json:"listing_type_label"`
	PropertyType      string     `json:"property_type"`
	PropertyTypeLabel string     `json:"property_type_label"`
	City              string     `json:"city"`
	Province          string     `json:"province"`
	Price             int64      `json:"price"`
	PriceLabel        string     `json:"price_label"`
	PricePeriod       string     `json:"price_period,omitempty"`
	LandArea          int        `json:"land_area"`
	BuildingArea      int        `json:"building_area"`
	Bedrooms          int        `json:"bedrooms"`
	Bathrooms         int        `json:"bathrooms"`
	UnitsAvailable    int        `json:"units_available"`
	CoverURL          string     `json:"cover_url"`
	Status            string     `json:"status"`
	StatusLabel       string     `json:"status_label"`
	Views             int        `json:"views"`
	IsFavorite        bool       `json:"is_favorite"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// NewPropertySummary membentuk ringkasan properti dari entitas domain.
func NewPropertySummary(property *models.Property) PropertySummary {
	if property == nil {
		return PropertySummary{}
	}
	return PropertySummary{
		ID:                property.ID,
		Slug:              property.Slug,
		Title:             property.Title,
		ListingType:       string(property.ListingType),
		ListingTypeLabel:  listingTypeLabel(property.ListingType),
		PropertyType:      string(property.PropertyType),
		PropertyTypeLabel: propertyTypeLabel(property.PropertyType),
		City:              property.City,
		Province:          property.Province,
		Price:             property.Price,
		PriceLabel:        priceLabel(property.Price, property.PricePeriod),
		PricePeriod:       string(property.PricePeriod),
		LandArea:          property.LandArea,
		BuildingArea:      property.BuildingArea,
		Bedrooms:          property.Bedrooms,
		Bathrooms:         property.Bathrooms,
		UnitsAvailable:    property.UnitsAvailable,
		CoverURL:          property.CoverURL(),
		Status:            string(property.Status),
		StatusLabel:       property.Status.Label(),
		Views:             property.Views,
		PublishedAt:       property.PublishedAt,
		CreatedAt:         property.CreatedAt,
	}
}

// PropertyDetail adalah bentuk lengkap properti untuk halaman detail dan form.
type PropertyDetail struct {
	PropertySummary
	Description  string          `json:"description"`
	Address      string          `json:"address"`
	District     string          `json:"district"`
	Latitude     float64         `json:"latitude"`
	Longitude    float64         `json:"longitude"`
	Facilities   []string        `json:"facilities"`
	UnitsTotal   int             `json:"units_total"`
	ContactCount int             `json:"contact_count"`
	PartnerID    string          `json:"partner_id"`
	PartnerName  string          `json:"partner_name"`
	Partner      *PartnerSummary `json:"partner,omitempty"`
	Photos       []PhotoView     `json:"photos"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// NewPropertyDetail membentuk detail properti dari entitas domain.
func NewPropertyDetail(property *models.Property) PropertyDetail {
	if property == nil {
		return PropertyDetail{}
	}
	facilities := property.Facilities
	if facilities == nil {
		facilities = []string{}
	}
	return PropertyDetail{
		PropertySummary: NewPropertySummary(property),
		Description:     property.Description,
		Address:         property.Address,
		District:        property.District,
		Latitude:        property.Latitude,
		Longitude:       property.Longitude,
		Facilities:      facilities,
		UnitsTotal:      property.UnitsTotal,
		ContactCount:    property.ContactCount,
		PartnerID:       property.PartnerID,
		PartnerName:     property.PartnerName,
		Photos:          NewPhotoViews(property.Photos),
		DeletedAt:       property.DeletedAt,
		UpdatedAt:       property.UpdatedAt,
	}
}

// NewPropertySummaries membentuk daftar ringkasan properti.
func NewPropertySummaries(properties []*models.Property) []PropertySummary {
	views := make([]PropertySummary, 0, len(properties))
	for _, property := range properties {
		views = append(views, NewPropertySummary(property))
	}
	return views
}

// PropertyFilterQuery adalah parameter pencarian katalog publik dan tabel properti.
type PropertyFilterQuery struct {
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page"`
	Search       string `form:"search"`
	City         string `form:"city"`
	Province     string `form:"province"`
	ListingType  string `form:"listing_type" validate:"omitempty,oneof=dijual disewa"`
	PropertyType string `form:"property_type" validate:"omitempty,oneof=rumah apartemen ruko tanah villa kantor gudang lainnya"`
	Status       string `form:"status"`
	PriceMin     int64  `form:"price_min" validate:"omitempty,gte=0"`
	PriceMax     int64  `form:"price_max" validate:"omitempty,gte=0"`
	SortBy       string `form:"sort_by" validate:"omitempty,oneof=price views title created_at"`
	SortDesc     bool   `form:"sort_desc"`
}

// ToListQuery mengubah filter properti menjadi ListQuery repository.
func (q PropertyFilterQuery) ToListQuery() database.ListQuery {
	filters := map[string]string{}
	if q.City != "" {
		filters["city"] = q.City
	}
	if q.Province != "" {
		filters["province"] = q.Province
	}
	if q.ListingType != "" {
		filters["listing_type"] = q.ListingType
	}
	if q.PropertyType != "" {
		filters["property_type"] = q.PropertyType
	}
	if q.Status != "" {
		filters["status"] = q.Status
	}
	if q.PriceMin > 0 {
		filters["price_min"] = strconv.FormatInt(q.PriceMin, 10)
	}
	if q.PriceMax > 0 {
		filters["price_max"] = strconv.FormatInt(q.PriceMax, 10)
	}
	return database.ListQuery{
		Search:   q.Search,
		Page:     q.Page,
		PerPage:  q.PerPage,
		SortBy:   q.SortBy,
		SortDesc: q.SortDesc,
		Filters:  filters,
	}
}

// CreatePropertyRequest adalah body POST /vendor/properties dan form mitra.
type CreatePropertyRequest struct {
	Title        string   `json:"title" validate:"required,min=5,max=150"`
	Description  string   `json:"description" validate:"required,min=20,max=5000"`
	ListingType  string   `json:"listing_type" validate:"required,oneof=dijual disewa"`
	PropertyType string   `json:"property_type" validate:"required,oneof=rumah apartemen ruko tanah villa kantor gudang lainnya"`
	Address      string   `json:"address" validate:"required,max=255"`
	City         string   `json:"city" validate:"required,max=100"`
	Province     string   `json:"province" validate:"required,max=100"`
	District     string   `json:"district" validate:"omitempty,max=100"`
	Latitude     float64  `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude    float64  `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	Price        int64    `json:"price" validate:"required,gte=0"`
	PricePeriod  string   `json:"price_period" validate:"omitempty,oneof=hari bulan tahun"`
	LandArea     int      `json:"land_area" validate:"omitempty,gte=0"`
	BuildingArea int      `json:"building_area" validate:"omitempty,gte=0"`
	Bedrooms     int      `json:"bedrooms" validate:"omitempty,gte=0,lte=50"`
	Bathrooms    int      `json:"bathrooms" validate:"omitempty,gte=0,lte=50"`
	Facilities   []string `json:"facilities" validate:"omitempty,dive,max=60"`
	UnitsTotal   int      `json:"units_total" validate:"omitempty,gte=0,lte=10000"`
	// Status draft menyimpan sebagai draf, active langsung menayangkan.
	Status string `json:"status" validate:"omitempty,oneof=draft active"`
}

// SetPropertyStatusRequest adalah body PATCH /vendor/properties/{id}/status.
type SetPropertyStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active inactive"`
}

// SetAvailabilityRequest adalah body PATCH /vendor/properties/{id}/availability.
type SetAvailabilityRequest struct {
	UnitsAvailable int `json:"units_available" validate:"gte=0,lte=10000"`
}

// AdminPropertyStatusRequest adalah body PATCH /admin/properties/{id}/status.
type AdminPropertyStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft pending_review active inactive expired"`
	Reason string `json:"reason" validate:"omitempty,max=255"`
}

// AdminCreatePropertyRequest adalah body POST /admin/properties untuk mitra Free.
type AdminCreatePropertyRequest struct {
	PartnerID string `json:"partner_id" validate:"required"`
	CreatePropertyRequest
	// Publish langsung menayangkan properti setelah dibuat.
	Publish bool `json:"publish"`
}

// AdminUpdatePropertyRequest adalah body PUT /admin/properties/{id}.
type AdminUpdatePropertyRequest struct {
	Title          *string  `json:"title" validate:"omitempty,min=5,max=150"`
	Description    *string  `json:"description" validate:"omitempty,min=20,max=5000"`
	ListingType    *string  `json:"listing_type" validate:"omitempty,oneof=dijual disewa"`
	PropertyType   *string  `json:"property_type" validate:"omitempty,oneof=rumah apartemen ruko tanah villa kantor gudang lainnya"`
	Address        *string  `json:"address" validate:"omitempty,max=255"`
	City           *string  `json:"city" validate:"omitempty,max=100"`
	Province       *string  `json:"province" validate:"omitempty,max=100"`
	District       *string  `json:"district" validate:"omitempty,max=100"`
	Latitude       *float64 `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude      *float64 `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	Price          *int64   `json:"price" validate:"omitempty,gte=0"`
	PricePeriod    *string  `json:"price_period" validate:"omitempty,oneof=hari bulan tahun"`
	LandArea       *int     `json:"land_area" validate:"omitempty,gte=0"`
	BuildingArea   *int     `json:"building_area" validate:"omitempty,gte=0"`
	Bedrooms       *int     `json:"bedrooms" validate:"omitempty,gte=0,lte=50"`
	Bathrooms      *int     `json:"bathrooms" validate:"omitempty,gte=0,lte=50"`
	Facilities     []string `json:"facilities" validate:"omitempty,dive,max=60"`
	UnitsTotal     *int     `json:"units_total" validate:"omitempty,gte=0,lte=10000"`
	UnitsAvailable *int     `json:"units_available" validate:"omitempty,gte=0,lte=10000"`
	Status         *string  `json:"status" validate:"omitempty,oneof=draft pending_review active inactive expired"`
}

// priceLabel menambahkan satuan periode pada label harga sewa.
func priceLabel(price int64, period models.PricePeriod) string {
	if period == "" {
		return money.Rupiah(price)
	}
	return money.Rupiah(price) + " / " + string(period)
}

// listingTypeLabel mengembalikan label jenis penawaran dalam bahasa Indonesia.
func listingTypeLabel(t models.ListingType) string {
	if t == models.ListingTypeRent {
		return "Disewa"
	}
	return "Dijual"
}

// propertyTypeLabel mengembalikan label jenis properti dengan huruf awal besar.
func propertyTypeLabel(t models.PropertyType) string {
	switch t {
	case models.PropertyTypeHouse:
		return "Rumah"
	case models.PropertyTypeApartment:
		return "Apartemen"
	case models.PropertyTypeRuko:
		return "Ruko"
	case models.PropertyTypeLand:
		return "Tanah"
	case models.PropertyTypeVilla:
		return "Villa"
	case models.PropertyTypeOffice:
		return "Kantor"
	case models.PropertyTypeWarehouse:
		return "Gudang"
	case models.PropertyTypeOther:
		return "Lainnya"
	default:
		return string(t)
	}
}
