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
type Plan struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
	DurationDays int       `json:"duration_days"`
	ListingQuota int       `json:"listing_quota"`
	Features     []string  `json:"features"`
	IsFree       bool      `json:"is_free"`
	IsActive     bool      `json:"is_active"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsUnlimited memeriksa paket dengan kuota listing tanpa batas.
func (p *Plan) IsUnlimited() bool { return p.ListingQuota < 0 }
