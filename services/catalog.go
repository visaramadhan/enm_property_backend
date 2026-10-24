package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/pkg/phone"
	"enmproperty/backend/platform/httpx"
)

// CatalogService melayani katalog publik: pencarian, detail properti, dan nomor
// kontak WhatsApp mitra atau admin (BE-11).
type CatalogService struct {
	deps     Deps
	settings *SettingService
}

// Search mengembalikan properti tayang sesuai filter katalog publik.
func (s *CatalogService) Search(q dto.PropertyFilterQuery) (dto.ListResult[dto.PropertySummary], error) {
	query := q.ToListQuery()
	if query.Filters == nil {
		query.Filters = newFilters()
	}
	query.Filters["public"] = "true"

	properties, total, err := s.deps.Repos.Properties.List(query)
	if err != nil {
		return dto.ListResult[dto.PropertySummary]{}, httpx.Internal(err)
	}
	return dto.NewListResult(dto.NewPropertySummaries(properties), total), nil
}

// GetDetail mengembalikan detail properti tayang dan mencatat tayangannya untuk
// statistik dashboard mitra.
func (s *CatalogService) GetDetail(slugOrID string, viewer *models.User) (dto.PropertyDetail, error) {
	property, err := s.publicProperty(slugOrID)
	if err != nil {
		return dto.PropertyDetail{}, err
	}
	s.recordView(property)

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
	return view, nil
}

// GetContact mengembalikan nomor WhatsApp mitra, atau nomor admin sebagai
// cadangan bila mitra belum mengisi kontaknya.
func (s *CatalogService) GetContact(slugOrID string, viewer *models.User) (dto.ContactResponse, error) {
	property, err := s.publicProperty(slugOrID)
	if err != nil {
		return dto.ContactResponse{}, err
	}

	response := dto.ContactResponse{
		PropertyID:    property.ID,
		PropertyTitle: property.Title,
		PartnerName:   property.PartnerName,
	}

	number := ""
	if partner, err := s.deps.Repos.Partners.GetByID(property.PartnerID); err == nil && partner != nil {
		response.PartnerName = partner.BusinessName
		number = partner.Phone
	}
	if strings.TrimSpace(number) == "" {
		number = s.settings.ContactFallback()
		response.IsFallback = true
	}

	response.Phone = number
	response.Message = s.contactMessage(property)
	response.WhatsAppURL = phone.WaMeURL(number, response.Message)

	s.recordContact(property, viewer)
	return response, nil
}

// publicProperty mengambil properti yang boleh tampil di katalog publik.
func (s *CatalogService) publicProperty(slugOrID string) (*models.Property, error) {
	key := strings.TrimSpace(slugOrID)
	if key == "" {
		return nil, httpx.BadRequest("Properti tidak disebutkan")
	}

	property, err := s.deps.Repos.Properties.GetBySlug(key)
	if err != nil {
		property, err = s.deps.Repos.Properties.GetByID(key)
	}
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Properti tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	if !property.IsPublicVisible() {
		return nil, httpx.NotFound("Properti tidak ditemukan")
	}
	return property, nil
}

// contactMessage menyusun pesan WhatsApp awal berisi judul dan tautan properti.
func (s *CatalogService) contactMessage(property *models.Property) string {
	base := strings.TrimRight(s.deps.Config.App.FrontendURL, "/")
	link := base + "/properti/" + property.Slug
	return "Halo, saya tertarik dengan properti \"" + property.Title + "\" (" + link + "). Mohon informasi lebih lanjut."
}

// recordView mencatat satu tayangan halaman detail properti.
func (s *CatalogService) recordView(property *models.Property) {
	_ = s.deps.Repos.Analytics.AddView(&models.PropertyView{
		ID:         newID(),
		PropertyID: property.ID,
		PartnerID:  property.PartnerID,
		ViewedAt:   timeNow(),
	})
}

// recordContact mencatat satu lead WhatsApp yang diteruskan ke mitra.
func (s *CatalogService) recordContact(property *models.Property, viewer *models.User) {
	contact := &models.PropertyContact{
		ID:         newID(),
		PropertyID: property.ID,
		PartnerID:  property.PartnerID,
		Channel:    "whatsapp",
		CreatedAt:  timeNow(),
	}
	if viewer != nil {
		contact.UserID = viewer.ID
	}
	_ = s.deps.Repos.Analytics.AddContact(contact)
}
