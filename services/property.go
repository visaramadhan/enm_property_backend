package services

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/slug"
	"enmproperty/backend/platform/httpx"
)

// contentPolicy membersihkan HTML dari konten yang dikirim pengguna agar aman
// ditampilkan kembali (cegah XSS).
var contentPolicy = bluemonday.UGCPolicy()

// sanitize membuang tag HTML berbahaya dari teks deskripsi.
func sanitize(html string) string { return strings.TrimSpace(contentPolicy.Sanitize(html)) }

// PropertyService mengelola listing properti mitra maupun admin (BE-09).
type PropertyService struct {
	deps          Deps
	subscriptions *SubscriptionService
}

// EnforceSlot memastikan mitra masih memiliki paket aktif dan kuota listing tersisa.
func (s *PropertyService) EnforceSlot(partnerID string) error {
	subscription := s.subscriptions.Active(partnerID)
	if subscription == nil || !subscription.IsActiveAt(timeNow()) {
		return httpx.Unprocessable("Mitra belum memiliki paket aktif, silakan pilih paket terlebih dahulu")
	}

	plan, err := s.deps.Repos.Plans.GetByID(subscription.PlanID)
	if err != nil {
		return nil
	}
	if plan.IsUnlimited() {
		return nil
	}
	if s.deps.Repos.Properties.CountByPartner(partnerID) >= plan.ListingQuota {
		return httpx.Unprocessable("Kuota listing paket sudah penuh, silakan upgrade paket Anda")
	}
	return nil
}

// Create menambahkan properti baru milik mitra.
func (s *PropertyService) Create(partner *models.Partner, req dto.CreatePropertyRequest) (dto.PropertyDetail, error) {
	if partner == nil {
		return dto.PropertyDetail{}, httpx.BadRequest("Profil kemitraan tidak tersedia")
	}
	if err := s.EnforceSlot(partner.ID); err != nil {
		return dto.PropertyDetail{}, err
	}

	property, err := s.build(partner, req, "")
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	if err := s.deps.Repos.Properties.Create(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// Update mengubah data properti milik mitra.
func (s *PropertyService) Update(partner *models.Partner, id string, req dto.CreatePropertyRequest) (dto.PropertyDetail, error) {
	property, err := s.owned(partner, id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}

	rebuilt, err := s.build(partner, req, property.ID)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	applyPropertyFields(property, rebuilt)

	if req.Status != "" {
		property.Status = models.PropertyStatus(req.Status)
		if property.Status == models.PropertyStatusActive && property.PublishedAt == nil {
			now := timeNow()
			property.PublishedAt = &now
		}
	}
	if property.UnitsAvailable > property.UnitsTotal {
		property.UnitsAvailable = property.UnitsTotal
	}
	property.UpdatedAt = timeNow()

	if err := s.deps.Repos.Properties.Update(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// Delete melakukan penghapusan lunak properti milik mitra.
func (s *PropertyService) Delete(partner *models.Partner, id string) error {
	if _, err := s.owned(partner, id); err != nil {
		return err
	}
	if err := s.deps.Repos.Properties.Delete(id); err != nil {
		if database.IsNotFound(err) {
			return httpx.NotFound("Properti tidak ditemukan")
		}
		return httpx.Internal(err)
	}
	return nil
}

// ListByPartner mengembalikan daftar properti milik mitra.
func (s *PropertyService) ListByPartner(partner *models.Partner, q dto.PropertyFilterQuery) (dto.ListResult[dto.PropertySummary], error) {
	if partner == nil {
		return dto.ListResult[dto.PropertySummary]{}, httpx.BadRequest("Profil kemitraan tidak tersedia")
	}
	query := q.ToListQuery()
	if query.Filters == nil {
		query.Filters = newFilters()
	}
	query.Filters["partner_id"] = partner.ID

	properties, total, err := s.deps.Repos.Properties.List(query)
	if err != nil {
		return dto.ListResult[dto.PropertySummary]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewPropertySummaries(properties), total), nil
}

// ListForAdmin mengembalikan daftar properti seluruh mitra untuk halaman admin.
func (s *PropertyService) ListForAdmin(q dto.PropertyFilterQuery) (dto.ListResult[dto.PropertySummary], error) {
	properties, total, err := s.deps.Repos.Properties.List(q.ToListQuery())
	if err != nil {
		return dto.ListResult[dto.PropertySummary]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewPropertySummaries(properties), total), nil
}

// Get mengembalikan detail properti berdasarkan ID.
func (s *PropertyService) Get(id string) (dto.PropertyDetail, error) {
	property, err := s.GetEntity(id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	return s.detail(property), nil
}

// GetEntity mengambil entitas properti berdasarkan ID.
func (s *PropertyService) GetEntity(id string) (*models.Property, error) {
	property, err := s.deps.Repos.Properties.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Properti tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return property, nil
}

// SetStatus mengubah status tayang properti milik mitra.
func (s *PropertyService) SetStatus(partner *models.Partner, id string, status models.PropertyStatus) (dto.PropertyDetail, error) {
	property, err := s.owned(partner, id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	return s.applyStatus(property, status)
}

// SetStatusByAdmin mengubah status properti oleh pengelola platform.
func (s *PropertyService) SetStatusByAdmin(id string, status models.PropertyStatus, reason string) (dto.PropertyDetail, error) {
	_ = reason
	property, err := s.GetEntity(id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	return s.applyStatus(property, status)
}

// SetAvailability memperbarui jumlah unit yang masih tersedia.
func (s *PropertyService) SetAvailability(partner *models.Partner, id string, units int) (dto.PropertyDetail, error) {
	property, err := s.owned(partner, id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	if units > property.UnitsTotal {
		return dto.PropertyDetail{}, httpx.BadRequest("Jumlah unit tersedia tidak boleh melebihi total unit")
	}
	property.UnitsAvailable = units
	property.UpdatedAt = timeNow()
	if err := s.deps.Repos.Properties.Update(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// AdminCreateForFree menambahkan properti atas nama mitra paket Free.
func (s *PropertyService) AdminCreateForFree(req dto.AdminCreatePropertyRequest) (dto.PropertyDetail, error) {
	partner, err := s.deps.Repos.Partners.GetByID(req.PartnerID)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.PropertyDetail{}, httpx.NotFound("Mitra tidak ditemukan")
		}
		return dto.PropertyDetail{}, httpx.Internal(err)
	}

	property, err := s.build(partner, req.CreatePropertyRequest, "")
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	if req.Publish {
		now := timeNow()
		property.Status = models.PropertyStatusActive
		property.PublishedAt = &now
	}
	if err := s.deps.Repos.Properties.Create(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// AdminUpdate mengubah data properti mitra oleh pengelola platform.
func (s *PropertyService) AdminUpdate(id string, req dto.AdminUpdatePropertyRequest) (dto.PropertyDetail, error) {
	property, err := s.GetEntity(id)
	if err != nil {
		return dto.PropertyDetail{}, err
	}

	if req.Title != nil {
		property.Title = strings.TrimSpace(*req.Title)
		property.Slug = s.uniqueSlug(property.Title, property.ID)
	}
	if req.Description != nil {
		property.Description = sanitize(*req.Description)
	}
	if req.ListingType != nil {
		property.ListingType = models.ListingType(*req.ListingType)
	}
	if req.PropertyType != nil {
		property.PropertyType = models.PropertyType(*req.PropertyType)
	}
	if req.Address != nil {
		property.Address = strings.TrimSpace(*req.Address)
	}
	if req.City != nil {
		property.City = strings.TrimSpace(*req.City)
	}
	if req.Province != nil {
		property.Province = strings.TrimSpace(*req.Province)
	}
	if req.District != nil {
		property.District = strings.TrimSpace(*req.District)
	}
	if req.Latitude != nil {
		property.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		property.Longitude = *req.Longitude
	}
	if req.Price != nil {
		property.Price = *req.Price
	}
	if req.PricePeriod != nil {
		property.PricePeriod = models.PricePeriod(*req.PricePeriod)
	}
	if req.LandArea != nil {
		property.LandArea = *req.LandArea
	}
	if req.BuildingArea != nil {
		property.BuildingArea = *req.BuildingArea
	}
	if req.Bedrooms != nil {
		property.Bedrooms = *req.Bedrooms
	}
	if req.Bathrooms != nil {
		property.Bathrooms = *req.Bathrooms
	}
	if req.Facilities != nil {
		property.Facilities = cleanList(req.Facilities)
	}
	if req.UnitsTotal != nil {
		property.UnitsTotal = *req.UnitsTotal
		if property.UnitsAvailable > property.UnitsTotal {
			property.UnitsAvailable = property.UnitsTotal
		}
	}
	if req.UnitsAvailable != nil {
		property.UnitsAvailable = *req.UnitsAvailable
	}
	if req.Status != nil {
		property.Status = models.PropertyStatus(*req.Status)
		if property.Status == models.PropertyStatusActive && property.PublishedAt == nil {
			now := timeNow()
			property.PublishedAt = &now
		}
	}
	property.UpdatedAt = timeNow()

	if err := s.deps.Repos.Properties.Update(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// applyStatus menyimpan perubahan status properti.
func (s *PropertyService) applyStatus(property *models.Property, status models.PropertyStatus) (dto.PropertyDetail, error) {
	now := timeNow()
	property.Status = status
	if status == models.PropertyStatusActive && property.PublishedAt == nil {
		property.PublishedAt = &now
	}
	property.UpdatedAt = now

	if err := s.deps.Repos.Properties.Update(property); err != nil {
		return dto.PropertyDetail{}, httpx.Internal(err)
	}
	return s.detail(property), nil
}

// build menyusun entitas properti dari permintaan pembuatan mitra.
func (s *PropertyService) build(partner *models.Partner, req dto.CreatePropertyRequest, excludeID string) (*models.Property, error) {
	listingType := models.ListingType(req.ListingType)
	if !listingType.Valid() {
		return nil, httpx.BadRequest("Jenis penawaran tidak dikenal")
	}
	propertyType := models.PropertyType(req.PropertyType)
	if !propertyType.Valid() {
		return nil, httpx.BadRequest("Jenis properti tidak dikenal")
	}
	period := models.PricePeriod(req.PricePeriod)
	if period != "" && !period.Valid() {
		return nil, httpx.BadRequest("Periode harga tidak dikenal")
	}
	if listingType != models.ListingTypeRent {
		period = ""
	}

	now := timeNow()
	status := models.PropertyStatusDraft
	if req.Status == string(models.PropertyStatusActive) {
		status = models.PropertyStatusActive
	}

	property := &models.Property{
		ID:             newID(),
		PartnerID:      partner.ID,
		PartnerName:    partner.BusinessName,
		Slug:           s.uniqueSlug(req.Title, excludeID),
		Title:          strings.TrimSpace(req.Title),
		Description:    sanitize(req.Description),
		ListingType:    listingType,
		PropertyType:   propertyType,
		Address:        strings.TrimSpace(req.Address),
		City:           strings.TrimSpace(req.City),
		Province:       strings.TrimSpace(req.Province),
		District:       strings.TrimSpace(req.District),
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Price:          req.Price,
		PricePeriod:    period,
		LandArea:       req.LandArea,
		BuildingArea:   req.BuildingArea,
		Bedrooms:       req.Bedrooms,
		Bathrooms:      req.Bathrooms,
		Facilities:     cleanList(req.Facilities),
		UnitsTotal:     req.UnitsTotal,
		UnitsAvailable: req.UnitsTotal,
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if status == models.PropertyStatusActive {
		property.PublishedAt = &now
	}
	return property, nil
}

// owned mengambil properti dan memastikan kepemilikannya oleh mitra.
func (s *PropertyService) owned(partner *models.Partner, id string) (*models.Property, error) {
	if partner == nil {
		return nil, httpx.BadRequest("Profil kemitraan tidak tersedia")
	}
	property, err := s.GetEntity(id)
	if err != nil {
		return nil, err
	}
	if property.PartnerID != partner.ID {
		return nil, httpx.Forbidden("Anda tidak berhak mengubah properti ini")
	}
	return property, nil
}

// uniqueSlug memastikan slug properti unik dengan menambahkan akhiran numerik.
func (s *PropertyService) uniqueSlug(title, excludeID string) string {
	base := slug.Make(title)
	if base == "" {
		base = "properti"
	}
	candidate := base
	for seq := 2; s.deps.Repos.Properties.SlugExists(candidate, excludeID); seq++ {
		candidate = slug.MakeWithSuffix(title, seq)
	}
	return candidate
}

// detail membentuk detail properti beserta ringkasan mitranya.
func (s *PropertyService) detail(property *models.Property) dto.PropertyDetail {
	view := dto.NewPropertyDetail(property)
	if partner, err := s.deps.Repos.Partners.GetByID(property.PartnerID); err == nil && partner != nil {
		view.Partner = &dto.PartnerSummary{
			ID:           partner.ID,
			BusinessName: partner.BusinessName,
			PICName:      partner.PICName,
			Phone:        partner.Phone,
			Email:        partner.Email,
			City:         partner.City,
			Province:     partner.Province,
		}
		view.PartnerName = partner.BusinessName
	}
	return view
}

// applyPropertyFields menyalin bidang yang dapat diubah dari satu entitas ke entitas lain.
func applyPropertyFields(dst, src *models.Property) {
	dst.Slug = src.Slug
	dst.Title = src.Title
	dst.Description = src.Description
	dst.ListingType = src.ListingType
	dst.PropertyType = src.PropertyType
	dst.Address = src.Address
	dst.City = src.City
	dst.Province = src.Province
	dst.District = src.District
	dst.Latitude = src.Latitude
	dst.Longitude = src.Longitude
	dst.Price = src.Price
	dst.PricePeriod = src.PricePeriod
	dst.LandArea = src.LandArea
	dst.BuildingArea = src.BuildingArea
	dst.Bedrooms = src.Bedrooms
	dst.Bathrooms = src.Bathrooms
	dst.Facilities = src.Facilities
	dst.UnitsTotal = src.UnitsTotal
	dst.UnitsAvailable = src.UnitsAvailable
	dst.PartnerName = src.PartnerName
}
