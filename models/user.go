package models

import "time"

// User adalah akun pengguna yang disinkronkan dari Firebase Authentication.
type User struct {
	ID            string     `json:"id"`
	FirebaseUID   string     `json:"firebase_uid"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	AvatarURL     string     `json:"avatar_url"`
	Role          Role       `json:"role"`
	Status        UserStatus `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// IsActive memeriksa apakah akun boleh mengakses API.
func (u *User) IsActive() bool { return u.Status != UserStatusDisabled }

// DisplayName mengembalikan nama tampilan dengan fallback ke email.
func (u *User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

// Partner adalah profil kemitraan yang dimiliki pengguna berperan mitra.
type Partner struct {
	ID                  string        `json:"id"`
	UserID              string        `json:"user_id"`
	BusinessName        string        `json:"business_name"`
	PICName             string        `json:"pic_name"`
	Phone               string        `json:"phone"`
	Email               string        `json:"email"`
	Address             string        `json:"address"`
	City                string        `json:"city"`
	Province            string        `json:"province"`
	Status              PartnerStatus `json:"status"`
	SuspendedReason     string        `json:"suspended_reason,omitempty"`
	AgreementAcceptedAt *time.Time    `json:"agreement_accepted_at,omitempty"`
	JoinedAt            *time.Time    `json:"joined_at,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// IsActive memeriksa apakah mitra boleh menulis properti.
func (p *Partner) IsActive() bool { return p.Status == PartnerStatusActive }

// Plan adalah paket langganan yang dijual pada halaman pricing.
//
// Selain harga dan kuota, paket mengatur hak akses mitra: level (dipakai untuk
// bobot boost di katalog), serta akses ke pengaturan keyword, katalog, dan
// nomor WhatsApp. Paket gratis tidak memperoleh ketiga akses tersebut.
type Plan struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        int64    `json:"price"`
	DurationDays int      `json:"duration_days"`
	ListingQuota int      `json:"listing_quota"`
	Features     []string `json:"features"`
	IsFree       bool     `json:"is_free"`
	IsActive     bool     `json:"is_active"`
	SortOrder    int      `json:"sort_order"`
	// Level menentukan prioritas tampil di katalog (1 terendah, makin tinggi makin atas).
	Level int `json:"level"`
	// BoostEnabled menandakan paket boleh memakai fitur boost listing.
	BoostEnabled bool `json:"boost_enabled"`
	// BoostQuota adalah jumlah listing yang boleh di-boost bersamaan. Nilai 0
	// berarti mitra tidak dapat memakai boost meski BoostEnabled bernilai true.
	BoostQuota int `json:"boost_quota"`
	// Akses pengaturan lanjutan sisi mitra.
	KeywordAccess  bool      `json:"keyword_access"`
	CatalogAccess  bool      `json:"catalog_access"`
	WhatsAppAccess bool      `json:"whatsapp_access"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IsUnlimited memeriksa paket dengan kuota listing tanpa batas.
func (p *Plan) IsUnlimited() bool { return p.ListingQuota < 0 }

// ApplyAccessRules menegakkan aturan akses sesuai sifat paket.
//
// Paket gratis (IsFree=true atau harga 0) tidak boleh memiliki akses pengaturan
// keyword, katalog, maupun nomor WhatsApp, dan tidak memperoleh boost. Level
// minimal 1 agar bobot pengurutan tetap terdefinisi.
func (p *Plan) ApplyAccessRules() {
	if p.Price == 0 {
		p.IsFree = true
	}
	if p.IsFree {
		p.KeywordAccess = false
		p.CatalogAccess = false
		p.WhatsAppAccess = false
		p.BoostEnabled = false
		p.BoostQuota = 0
	}
	if p.BoostQuota < 0 {
		p.BoostQuota = 0
	}
	if p.Level < 1 {
		p.Level = 1
	}
}
