package database

import (
	"sort"
	"strings"

	"enmproperty/backend/models"
)

// ErrNotFound dikembalikan repository ketika data tidak ditemukan.
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "data tidak ditemukan" }

// IsNotFound memeriksa apakah error berasal dari data yang tidak ditemukan.
func IsNotFound(err error) bool { return err == ErrNotFound }

// ListQuery adalah parameter daftar yang dipakai seluruh repository.
//
// PerPage bernilai 0 atau negatif berarti tanpa paginasi, dipakai modul laporan
// yang perlu mengagregasi seluruh data.
type ListQuery struct {
	Search   string
	Page     int
	PerPage  int
	SortBy   string
	SortDesc bool
	Filters  map[string]string
}

// Filter mengambil nilai filter berdasarkan kunci.
func (q ListQuery) Filter(key string) string {
	if q.Filters == nil {
		return ""
	}
	return q.Filters[key]
}

// Offset menghitung offset paginasi.
func (q ListQuery) Offset() int {
	if q.Page <= 1 || q.PerPage <= 0 {
		return 0
	}
	return (q.Page - 1) * q.PerPage
}

// Paginate memotong slice hasil sesuai paginasi dan sekaligus mengembalikan total.
func Paginate[T any](items []T, q ListQuery) ([]T, int) {
	total := len(items)
	if q.PerPage <= 0 {
		return items, total
	}
	offset := q.Offset()
	if offset >= total {
		return []T{}, total
	}
	end := offset + q.PerPage
	if end > total {
		end = total
	}
	return items[offset:end], total
}

// ContainsFold memeriksa apakah teks mengandung kata kunci tanpa memandang huruf besar-kecil.
func ContainsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// SortByFunc mengurutkan slice dengan fungsi komparator yang stabil.
func SortByFunc[T any](items []T, less func(a, b T) bool, desc bool) {
	sort.SliceStable(items, func(i, j int) bool {
		if desc {
			return less(items[j], items[i])
		}
		return less(items[i], items[j])
	})
}

// UserRepository menangani akses data pengguna.
type UserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByFirebaseUID(uid string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	List(q ListQuery) ([]*models.User, int, error)
}

// PartnerRepository menangani akses data kemitraan.
type PartnerRepository interface {
	Create(partner *models.Partner) error
	Update(partner *models.Partner) error
	GetByID(id string) (*models.Partner, error)
	GetByUserID(userID string) (*models.Partner, error)
	List(q ListQuery) ([]*models.Partner, int, error)
}

// PlanRepository menangani akses data paket langganan.
type PlanRepository interface {
	Create(plan *models.Plan) error
	Update(plan *models.Plan) error
	GetByID(id string) (*models.Plan, error)
	GetBySlug(slug string) (*models.Plan, error)
	List(q ListQuery) ([]*models.Plan, int, error)
}

// CategoryRepository menangani kategori properti untuk katalog dan master data.
type CategoryRepository interface {
	Create(category *models.Category) error
	Update(category *models.Category) error
	Delete(id string) error
	GetByID(id string) (*models.Category, error)
	GetBySlug(slug string) (*models.Category, error)
	SlugExists(slug, excludeID string) bool
	List(q ListQuery) ([]*models.Category, int, error)
}

// AgreementRepository menangani versi perjanjian dan bukti persetujuannya.
type AgreementRepository interface {
	Create(agreement *models.Agreement) error
	Update(agreement *models.Agreement) error
	GetByID(id string) (*models.Agreement, error)
	GetActive() (*models.Agreement, error)
	List(q ListQuery) ([]*models.Agreement, int, error)

	SaveAcceptance(acceptance *models.AgreementAcceptance) error
	GetAcceptance(partnerID, agreementID string) (*models.AgreementAcceptance, error)
	ListAcceptances(q ListQuery) ([]*models.AgreementAcceptance, int, error)
}

// SubscriptionRepository menangani masa aktif paket mitra.
type SubscriptionRepository interface {
	Create(subscription *models.Subscription) error
	Update(subscription *models.Subscription) error
	GetByID(id string) (*models.Subscription, error)
	GetActiveByPartner(partnerID string) (*models.Subscription, error)
	GetLatestByPartner(partnerID string) (*models.Subscription, error)
	List(q ListQuery) ([]*models.Subscription, int, error)
}

// OrderRepository menangani transaksi langganan.
type OrderRepository interface {
	Create(order *models.Order) error
	Update(order *models.Order) error
	GetByID(id string) (*models.Order, error)
	GetByCode(code string) (*models.Order, error)
	CountByCodePrefix(prefix string) int
	List(q ListQuery) ([]*models.Order, int, error)
}

// PropertyRepository menangani properti beserta fotonya.
type PropertyRepository interface {
	Create(property *models.Property) error
	Update(property *models.Property) error
	Delete(id string) error
	GetByID(id string) (*models.Property, error)
	GetBySlug(slug string) (*models.Property, error)
	SlugExists(slug, excludeID string) bool
	CountByPartner(partnerID string, statuses ...models.PropertyStatus) int
	UpdateStatusByPartner(partnerID string, from, to models.PropertyStatus) int
	List(q ListQuery) ([]*models.Property, int, error)

	CreatePhoto(photo *models.PropertyPhoto) error
	UpdatePhoto(photo *models.PropertyPhoto) error
	DeletePhoto(id string) error
	GetPhoto(id string) (*models.PropertyPhoto, error)
	ListPhotos(propertyID string) ([]*models.PropertyPhoto, error)
}

// FavoriteRepository menangani favorit customer.
type FavoriteRepository interface {
	Add(favorite *models.Favorite) error
	Remove(userID, propertyID string) error
	Exists(userID, propertyID string) bool
	ListByUser(userID string) ([]*models.Favorite, error)
	CountByProperty(propertyID string) int
}

// AnalyticsRepository menyimpan tayangan dan lead untuk laporan.
type AnalyticsRepository interface {
	AddView(view *models.PropertyView) error
	AddContact(contact *models.PropertyContact) error
	Views() []*models.PropertyView
	Contacts() []*models.PropertyContact
	// ListContactsByProperty mengembalikan lead pada satu properti (terbaru dulu).
	ListContactsByProperty(propertyID string) ([]*models.PropertyContact, error)
	// ContactsByPartner mengembalikan seluruh lead milik satu mitra (terbaru dulu).
	ContactsByPartner(partnerID string) ([]*models.PropertyContact, error)
}

// SettingRepository menangani pengaturan sistem.
type SettingRepository interface {
	All() []*models.Setting
	Get(key string) (*models.Setting, error)
	Set(setting *models.Setting) error
}

// LandingRepository menangani konten landing page (CMS).
type LandingRepository interface {
	All() []*models.LandingContent
	Get(key string) (*models.LandingContent, error)
	Set(content *models.LandingContent) error
}

// AuditRepository menangani catatan audit.
type AuditRepository interface {
	Create(entry *models.AuditLog) error
	List(q ListQuery) ([]*models.AuditLog, int, error)
	All() []*models.AuditLog
}

// Repositories mengumpulkan seluruh repository yang dibutuhkan services.
type Repositories struct {
	Users         UserRepository
	Partners      PartnerRepository
	Plans         PlanRepository
	Categories    CategoryRepository
	Agreements    AgreementRepository
	Subscriptions SubscriptionRepository
	Orders        OrderRepository
	Properties    PropertyRepository
	Favorites     FavoriteRepository
	Analytics     AnalyticsRepository
	Settings      SettingRepository
	Landings      LandingRepository
	Audits        AuditRepository
}

// NewMemoryRepositories membuat implementasi in-memory di atas Store.
func NewMemoryRepositories(store *Store) Repositories {
	return Repositories{
		Users:         &memoryUserRepository{store: store},
		Partners:      &memoryPartnerRepository{store: store},
		Plans:         &memoryPlanRepository{store: store},
		Categories:    &memoryCategoryRepository{store: store},
		Agreements:    &memoryAgreementRepository{store: store},
		Subscriptions: &memorySubscriptionRepository{store: store},
		Orders:        &memoryOrderRepository{store: store},
		Properties:    &memoryPropertyRepository{store: store},
		Favorites:     &memoryFavoriteRepository{store: store},
		Analytics:     &memoryAnalyticsRepository{store: store},
		Settings:      &memorySettingRepository{store: store},
		Landings:      &memoryLandingRepository{store: store},
		Audits:        &memoryAuditRepository{store: store},
	}
}
