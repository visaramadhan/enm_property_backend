// Package services memuat logika bisnis seluruh modul backend, dari BE-02 auth
// hingga BE-16 scheduler. Setiap service hanya bergantung pada repository
// (melalui database.Repositories) dan integrasi platform.
package services

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"enmproperty/backend/config"
	"enmproperty/backend/database"
	"enmproperty/backend/pkg/phone"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/midtrans"
	"enmproperty/backend/platform/storage"
)

// Deps mengumpulkan dependensi bersama seluruh service.
type Deps struct {
	Repos    database.Repositories
	Config   *config.Config
	Firebase *firebase.Client
	Midtrans *midtrans.Client
	Storage  storage.Storage
}

// RequestMeta membawa informasi permintaan yang dipakai untuk bukti persetujuan
// perjanjian dan audit manual.
type RequestMeta struct {
	IPAddress string
	UserAgent string
}

// Services mengumpulkan seluruh service per modul agar router cukup menerima
// satu nilai.
type Services struct {
	Auth         *AuthService
	User         *UserService
	Partner      *PartnerService
	Plan         *PlanService
	Category     *CategoryService
	Facility     *FacilityService
	Agreement    *AgreementService
	Subscription *SubscriptionService
	Payment      *PaymentService
	Property     *PropertyService
	Media        *MediaService
	Catalog      *CatalogService
	Favorite     *FavoriteService
	Report       *ReportService
	Setting      *SettingService
	Landing      *LandingService
	Audit        *AuditService
	Scheduler    *SchedulerService
}

// NewServices merakit seluruh service beserta ketergantungan antar modul.
func NewServices(deps Deps) *Services {
	setting := &SettingService{deps: deps}
	audit := &AuditService{deps: deps}
	agreement := &AgreementService{deps: deps}
	partner := &PartnerService{deps: deps, agreements: agreement}
	subscription := &SubscriptionService{deps: deps, settings: setting}
	property := &PropertyService{deps: deps, subscriptions: subscription}
	payment := &PaymentService{
		deps:          deps,
		partners:      partner,
		agreements:    agreement,
		subscriptions: subscription,
		settings:      setting,
	}
	catalog := &CatalogService{deps: deps, settings: setting}
	media := &MediaService{deps: deps}
	favorite := &FavoriteService{deps: deps}
	report := &ReportService{
		deps:          deps,
		subscriptions: subscription,
		properties:    property,
	}

	services := &Services{
		Auth:         &AuthService{deps: deps},
		User:         &UserService{deps: deps},
		Partner:      partner,
		Plan:         &PlanService{deps: deps},
		Category:     &CategoryService{deps: deps},
		Facility:     &FacilityService{deps: deps},
		Agreement:    agreement,
		Subscription: subscription,
		Payment:      payment,
		Property:     property,
		Media:        media,
		Catalog:      catalog,
		Favorite:     favorite,
		Report:       report,
		Setting:      setting,
		Landing:      &LandingService{deps: deps},
		Audit:        audit,
	}
	services.Scheduler = &SchedulerService{
		deps:          deps,
		subscriptions: subscription,
		payments:      payment,
		logger:        newSchedulerLogger(deps),
	}
	return services
}

// newFilters membuat peta filter kosong untuk repository.
func newFilters() map[string]string { return map[string]string{} }

// putFilter menambahkan filter hanya bila nilainya tidak kosong.
func putFilter(filters map[string]string, key, value string) {
	if strings.TrimSpace(value) != "" {
		filters[key] = value
	}
}

// newID membentuk UUID untuk entitas baru.
func newID() string { return uuid.NewString() }

// normalizePhone menormalkan nomor telepon ke format WhatsApp 62...
func normalizePhone(raw string) string { return phone.Normalize(raw) }

// firebasePhone melengkapi nomor dengan tanda plus sesuai syarat Firebase Admin.
func firebasePhone(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "+") {
		return trimmed
	}
	return "+" + trimmed
}

// displayNameFromEmail membentuk nama tampilan dari alamat email.
func displayNameFromEmail(email string) string {
	local, _, found := strings.Cut(email, "@")
	if !found || local == "" {
		return email
	}
	return local
}

// toJSON mengubah nilai menjadi teks JSON untuk kolom before/after audit.
func toJSON(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// firstNonEmpty mengembalikan nilai pertama yang tidak kosong.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// timeNow mengembalikan waktu sekarang, dipusatkan agar mudah diuji.
func timeNow() time.Time { return time.Now() }

// parseDays mengubah teks menjadi jumlah hari dengan nilai bawaan.
func parseDays(raw string, fallback int) int {
	value := 0
	for _, char := range strings.TrimSpace(raw) {
		if char < '0' || char > '9' {
			return fallback
		}
		value = value*10 + int(char-'0')
	}
	if value <= 0 {
		return fallback
	}
	return value
}
