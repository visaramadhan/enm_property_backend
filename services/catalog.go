package services

import (
	"math"
	"sort"
	"strings"
	"time"

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
//
// Seluruh hasil yang cocok diambil tanpa paginasi lebih dulu agar algoritma
// pengurutan (boost -> keyword -> klik WhatsApp -> rating -> favorit) dapat
// diterapkan sebelum halaman dipotong.
func (s *CatalogService) Search(q dto.PropertyFilterQuery, viewer *models.User) (dto.ListResult[dto.PropertySummary], error) {
	query := q.ToListQuery()
	if query.Filters == nil {
		query.Filters = newFilters()
	}
	query.Filters["public"] = "true"

	page, perPage := query.Page, query.PerPage
	query.Page, query.PerPage = 0, 0

	properties, _, err := s.deps.Repos.Properties.List(query)
	if err != nil {
		return dto.ListResult[dto.PropertySummary]{}, httpx.Internal(err)
	}

	s.rank(properties, q.Search)

	total := len(properties)
	pageItems := paginateProperties(properties, page, perPage)

	summaries := dto.NewPropertySummaries(pageItems)
	s.attachCategoryNames(summaries)
	s.markFavorites(summaries, viewer)
	return dto.NewListResult(summaries, total), nil
}

// markFavorites menandai ringkasan properti yang sudah tersimpan pada daftar
// favorit pengunjung yang sedang masuk, agar ikon favorit di katalog sesuai
// keadaan sebenarnya. Pengunjung anonim dibiarkan tanpa penanda (semua false)
// supaya hasil tetap sama untuk semua pengunjung.
func (s *CatalogService) markFavorites(summaries []dto.PropertySummary, viewer *models.User) {
	if viewer == nil || len(summaries) == 0 {
		return
	}
	for i := range summaries {
		summaries[i].IsFavorite = s.deps.Repos.Favorites.Exists(viewer.ID, summaries[i].ID)
	}
}

// rankedProperty menyimpan komponen skor pengurutan sebuah properti.
type rankedProperty struct {
	property *models.Property
	boost    int
	keyword  int
	wa       int
	rating   float64
	favorite int
}

// rank mengurutkan properti sesuai prioritas: boost, keyword, klik WhatsApp,
// rating, lalu favorit. Tie-break: terbit terbaru, lalu ID agar stabil.
func (s *CatalogService) rank(properties []*models.Property, keyword string) {
	subCache := map[string]*models.Subscription{}
	planCache := map[string]*models.Plan{}
	now := timeNow()

	levelFor := func(partnerID string) (int, bool) {
		sub, ok := subCache[partnerID]
		if !ok {
			found, err := s.deps.Repos.Subscriptions.GetActiveByPartner(partnerID)
			if err != nil {
				sub = nil
			} else {
				sub = found
			}
			subCache[partnerID] = sub
		}
		if sub == nil || !sub.IsActiveAt(now) {
			return 0, false
		}
		plan, ok := planCache[sub.PlanID]
		if !ok {
			found, err := s.deps.Repos.Plans.GetByID(sub.PlanID)
			if err != nil {
				plan = nil
			} else {
				plan = found
			}
			planCache[sub.PlanID] = plan
		}
		if plan == nil || !plan.BoostEnabled {
			return 0, false
		}
		return plan.Level, true
	}

	ranked := make([]rankedProperty, 0, len(properties))
	for _, property := range properties {
		item := rankedProperty{
			property: property,
			keyword:  keywordScore(property, keyword),
			wa:       property.ContactCount,
			rating:   ratingScore(property),
			favorite: s.favoriteCount(property),
		}
		// Boost hanya berlaku bila paket mitra mengaktifkan boost, listing
		// ditandai boost aktif, dan masa boost belum berakhir.
		if level, ok := levelFor(property.PartnerID); ok && property.BoostActive {
			if property.BoostUntil == nil || property.BoostUntil.After(now) {
				item.boost = level
			}
		}
		ranked = append(ranked, item)
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		if a.boost != b.boost {
			return a.boost > b.boost
		}
		if a.keyword != b.keyword {
			return a.keyword > b.keyword
		}
		if a.wa != b.wa {
			return a.wa > b.wa
		}
		if a.rating != b.rating {
			return a.rating > b.rating
		}
		if a.favorite != b.favorite {
			return a.favorite > b.favorite
		}
		return publishedAt(ranked[i].property).After(publishedAt(ranked[j].property))
	})

	for i := range ranked {
		properties[i] = ranked[i].property
	}
}

// keywordScore menilai kecocokan kata kunci pada judul, kota, alamat, dan deskripsi.
func keywordScore(property *models.Property, keyword string) int {
	k := strings.ToLower(strings.TrimSpace(keyword))
	if k == "" {
		return 0
	}
	score := 0
	if strings.Contains(strings.ToLower(property.Title), k) {
		score += 4
	}
	if strings.Contains(strings.ToLower(property.City), k) {
		score += 2
	}
	if strings.Contains(strings.ToLower(property.Address), k) {
		score += 1
	}
	if strings.Contains(strings.ToLower(property.Description), k) {
		score += 1
	}
	return score
}

// ratingScore menggabungkan rata-rata rating dengan jumlah ulasan agar properti
// dengan banyak ulasan positif lebih diutamakan.
func ratingScore(property *models.Property) float64 {
	if property.RatingCount <= 0 {
		return 0
	}
	return property.RatingAvg * math.Log(float64(property.RatingCount)+1)
}

// favoriteCount memakai jumlah favorit dari repository (sumber kebenaran),
// dengan fallback ke nilai denormalisasi pada properti.
func (s *CatalogService) favoriteCount(property *models.Property) int {
	count := s.deps.Repos.Favorites.CountByProperty(property.ID)
	if count > property.FavoriteCount {
		return count
	}
	return property.FavoriteCount
}

// attachCategoryNames melengkapi ringkasan properti dengan nama kategori.
func (s *CatalogService) attachCategoryNames(summaries []dto.PropertySummary) {
	if len(summaries) == 0 {
		return
	}
	categories, _, err := s.deps.Repos.Categories.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return
	}
	names := make(map[string]string, len(categories))
	for _, category := range categories {
		names[category.ID] = category.Name
	}
	for i := range summaries {
		if name, ok := names[summaries[i].CategoryID]; ok {
			summaries[i].CategoryName = name
		}
	}
}

// paginateProperties memotong hasil sesuai halaman; perPage <= 0 berarti tanpa batas.
func paginateProperties(items []*models.Property, page, perPage int) []*models.Property {
	if perPage <= 0 {
		return items
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * perPage
	}
	if offset >= len(items) {
		return []*models.Property{}
	}
	end := offset + perPage
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

// publishedAt mengembalikan waktu terbit, atau waktu pembuatan bila belum terbit.
func publishedAt(property *models.Property) time.Time {
	if property.PublishedAt != nil {
		return *property.PublishedAt
	}
	return property.CreatedAt
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
	if viewer != nil {
		// Status favorit bersifat per pengguna sehingga hanya diisi bila ada sesi.
		view.IsFavorite = s.deps.Repos.Favorites.Exists(viewer.ID, property.ID)
	}
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

// GetPartner mengembalikan profil pengelola (mitra) beserta katalog publik
// miliknya untuk halaman /pengelola/{id}.
func (s *CatalogService) GetPartner(id string) (dto.PartnerPublicView, error) {
	key := strings.TrimSpace(id)
	if key == "" {
		return dto.PartnerPublicView{}, httpx.BadRequest("Pengelola tidak disebutkan")
	}

	partner, err := s.deps.Repos.Partners.GetByID(key)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.PartnerPublicView{}, httpx.NotFound("Pengelola tidak ditemukan")
		}
		return dto.PartnerPublicView{}, httpx.Internal(err)
	}

	query := database.ListQuery{Filters: newFilters()}
	query.Filters["partner_id"] = partner.ID
	query.Filters["public"] = "true"

	properties, _, err := s.deps.Repos.Properties.List(query)
	if err != nil {
		return dto.PartnerPublicView{}, httpx.Internal(err)
	}
	s.rank(properties, "")

	summaries := dto.NewPropertySummaries(properties)
	s.attachCategoryNames(summaries)

	return dto.PartnerPublicView{
		ID:            partner.ID,
		BusinessName:  partner.BusinessName,
		PICName:       partner.PICName,
		City:          partner.City,
		Province:      partner.Province,
		JoinedAt:      partner.JoinedAt,
		TotalListings: len(summaries),
		Properties:    summaries,
	}, nil
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
	// Paket yang tidak memberi akses WhatsApp (mis. paket Gratis) selalu
	// memakai nomor sistem, meskipun mitra menyimpan nomor sendiri.
	if !s.whatsappDiizinkan(property.PartnerID) {
		number = ""
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

// SubmitContactForm mencatat lead dari form kontak publik pada halaman detail
// properti. Lead tersimpan dengan kanal "form" beserta data pengirim.
func (s *CatalogService) SubmitContactForm(slugOrID string, req dto.ContactFormRequest, viewer *models.User) (dto.ContactFormResponse, error) {
	property, err := s.publicProperty(slugOrID)
	if err != nil {
		return dto.ContactFormResponse{}, err
	}

	contact := &models.PropertyContact{
		ID:         newID(),
		PropertyID: property.ID,
		PartnerID:  property.PartnerID,
		Channel:    "form",
		Name:       strings.TrimSpace(req.Name),
		Phone:      strings.TrimSpace(req.Phone),
		Message:    strings.TrimSpace(req.Message),
		CreatedAt:  timeNow(),
	}
	if viewer != nil {
		contact.UserID = viewer.ID
	}

	if err := s.deps.Repos.Analytics.AddContact(contact); err != nil {
		return dto.ContactFormResponse{}, httpx.Internal(err)
	}
	return dto.ContactFormResponse{Recorded: true}, nil
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

// whatsappDiizinkan memeriksa apakah paket aktif mitra mengizinkan nomor
// WhatsApp sendiri. Paket tanpa akses (mis. paket Gratis) selalu memakai nomor
// sistem. Bila langganan atau paket tidak dapat dibaca, nomor mitra dipakai.
func (s *CatalogService) whatsappDiizinkan(partnerID string) bool {
	sub, err := s.deps.Repos.Subscriptions.GetActiveByPartner(partnerID)
	if err != nil || sub == nil || !sub.IsActiveAt(timeNow()) {
		return true
	}
	plan, err := s.deps.Repos.Plans.GetByID(sub.PlanID)
	if err != nil || plan == nil {
		return true
	}
	return plan.WhatsAppAccess
}

// contactMessage menyusun pesan WhatsApp awal berisi judul dan tautan properti.
//
// Tautan mengarah ke halaman detail katalog publik (`/katalog/<slug>`) agar
// mitra dapat membuka listing yang dimaksud dari pesan.
func (s *CatalogService) contactMessage(property *models.Property) string {
	base := strings.TrimRight(s.deps.Config.App.FrontendURL, "/")
	link := base + "/katalog/" + property.Slug
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
