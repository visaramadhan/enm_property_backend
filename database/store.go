// Package database menyediakan kontrak repository dan implementasi in-memory
// yang menjadi lapisan data default saat DB_DRIVER=memory.
//
// Struktur koleksi pada Store sengaja dibuat sejajar dengan tabel PostgreSQL
// agar implementasi pgx dapat ditambahkan kemudian tanpa mengubah kontrak
// Repositories yang sudah dipakai services.
package database

import (
	"sync"
	"time"

	"enmproperty/backend/models"
)

// Store adalah penyimpanan in-memory thread-safe untuk seluruh agregat.
type Store struct {
	mu sync.RWMutex

	users         map[string]*models.User
	partners      map[string]*models.Partner
	plans         map[string]*models.Plan
	categories    map[string]*models.Category
	agreements    map[string]*models.Agreement
	acceptances   map[string]*models.AgreementAcceptance
	subscriptions map[string]*models.Subscription
	orders        map[string]*models.Order
	properties    map[string]*models.Property
	photos        map[string]*models.PropertyPhoto
	favorites     map[string]*models.Favorite
	views         []*models.PropertyView
	contacts      []*models.PropertyContact
	settings      map[string]*models.Setting
	landing       map[string]*models.LandingContent
	audits        []*models.AuditLog
}

// NewStore membuat Store kosong dengan seluruh koleksi siap pakai.
func NewStore() *Store {
	return &Store{
		users:         make(map[string]*models.User),
		partners:      make(map[string]*models.Partner),
		plans:         make(map[string]*models.Plan),
		categories:    make(map[string]*models.Category),
		agreements:    make(map[string]*models.Agreement),
		acceptances:   make(map[string]*models.AgreementAcceptance),
		subscriptions: make(map[string]*models.Subscription),
		orders:        make(map[string]*models.Order),
		properties:    make(map[string]*models.Property),
		photos:        make(map[string]*models.PropertyPhoto),
		favorites:     make(map[string]*models.Favorite),
		settings:      make(map[string]*models.Setting),
		landing:       make(map[string]*models.LandingContent),
	}
}

// ---------- penyalinan ----------
//
// Repository selalu mengembalikan salinan agar pemanggil tidak dapat mengubah
// data tersimpan tanpa melalui method repository (menjaga konsistensi kunci).

func copyTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	value := *t
	return &value
}

func copyStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func copyUser(in *models.User) *models.User {
	if in == nil {
		return nil
	}
	out := *in
	out.LastLoginAt = copyTime(in.LastLoginAt)
	return &out
}

func copyPartner(in *models.Partner) *models.Partner {
	if in == nil {
		return nil
	}
	out := *in
	out.AgreementAcceptedAt = copyTime(in.AgreementAcceptedAt)
	out.JoinedAt = copyTime(in.JoinedAt)
	return &out
}

func copyPlan(in *models.Plan) *models.Plan {
	if in == nil {
		return nil
	}
	out := *in
	out.Features = copyStrings(in.Features)
	return &out
}

func copyCategory(in *models.Category) *models.Category {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyAgreement(in *models.Agreement) *models.Agreement {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyAcceptance(in *models.AgreementAcceptance) *models.AgreementAcceptance {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copySubscription(in *models.Subscription) *models.Subscription {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyOrder(in *models.Order) *models.Order {
	if in == nil {
		return nil
	}
	out := *in
	out.PaidAt = copyTime(in.PaidAt)
	out.PaymentExpiresAt = copyTime(in.PaymentExpiresAt)
	return &out
}

func copyPhoto(in *models.PropertyPhoto) *models.PropertyPhoto {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyProperty(in *models.Property) *models.Property {
	if in == nil {
		return nil
	}
	out := *in
	out.Facilities = copyStrings(in.Facilities)
	out.PublishedAt = copyTime(in.PublishedAt)
	out.DeletedAt = copyTime(in.DeletedAt)
	out.BoostUntil = copyTime(in.BoostUntil)
	if in.Photos != nil {
		out.Photos = make([]models.PropertyPhoto, len(in.Photos))
		for i := range in.Photos {
			out.Photos[i] = in.Photos[i]
		}
	}
	return &out
}

func copyFavorite(in *models.Favorite) *models.Favorite {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyView(in *models.PropertyView) *models.PropertyView {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyContact(in *models.PropertyContact) *models.PropertyContact {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copySetting(in *models.Setting) *models.Setting {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyLanding(in *models.LandingContent) *models.LandingContent {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func copyAudit(in *models.AuditLog) *models.AuditLog {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
