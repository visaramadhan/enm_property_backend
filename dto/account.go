package dto

import (
	"time"

	"enmproperty/backend/models"
)

// UserView adalah representasi pengguna untuk respons API.
type UserView struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Name          string       `json:"name"`
	Phone         string       `json:"phone"`
	AvatarURL     string       `json:"avatar_url"`
	Role          string       `json:"role"`
	RoleLabel     string       `json:"role_label"`
	Status        string       `json:"status"`
	EmailVerified bool         `json:"email_verified"`
	LastLoginAt   *time.Time   `json:"last_login_at,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Partner       *PartnerView `json:"partner,omitempty"`
}

// NewUserView membentuk tampilan pengguna dari entitas domain.
func NewUserView(user *models.User) UserView {
	if user == nil {
		return UserView{}
	}
	return UserView{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		Role:          string(user.Role),
		RoleLabel:     user.Role.Label(),
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

// WithPartner melampirkan profil mitra pada tampilan pengguna.
func (v UserView) WithPartner(partner *models.Partner) UserView {
	if partner != nil {
		view := NewPartnerView(partner)
		v.Partner = &view
	}
	return v
}

// SessionResponse adalah respons POST /auth/session.
type SessionResponse struct {
	User UserView `json:"user"`
	Role string   `json:"role"`
}

// UpdateMeRequest adalah body PUT /me.
type UpdateMeRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=100"`
	Phone string `json:"phone" validate:"omitempty,min=8,max=20"`
}

// AdminCreateUserRequest adalah body POST /admin/users.
type AdminCreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Phone    string `json:"phone" validate:"omitempty,min=8,max=20"`
	Role     string `json:"role" validate:"required,oneof=customer mitra admin superadmin owner"`
}

// AdminUpdateUserRequest adalah body PUT /admin/users/{id}.
type AdminUpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=3,max=100"`
	Phone string `json:"phone" validate:"omitempty,min=8,max=20"`
	Role  string `json:"role" validate:"omitempty,oneof=customer mitra admin superadmin owner"`
}

// SetUserStatusRequest adalah body PATCH /admin/users/{id}/status.
type SetUserStatusRequest struct {
	Active bool   `json:"active"`
	Reason string `json:"reason" validate:"omitempty,max=255"`
}

// PartnerView adalah representasi mitra untuk respons API.
type PartnerView struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	BusinessName        string     `json:"business_name"`
	PICName             string     `json:"pic_name"`
	Phone               string     `json:"phone"`
	Email               string     `json:"email"`
	Address             string     `json:"address"`
	City                string     `json:"city"`
	Province            string     `json:"province"`
	Status              string     `json:"status"`
	StatusLabel         string     `json:"status_label"`
	SuspendedReason     string     `json:"suspended_reason,omitempty"`
	AgreementAcceptedAt *time.Time `json:"agreement_accepted_at,omitempty"`
	JoinedAt            *time.Time `json:"joined_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// NewPartnerView membentuk tampilan mitra dari entitas domain.
func NewPartnerView(partner *models.Partner) PartnerView {
	if partner == nil {
		return PartnerView{}
	}
	return PartnerView{
		ID:                  partner.ID,
		UserID:              partner.UserID,
		BusinessName:        partner.BusinessName,
		PICName:             partner.PICName,
		Phone:               partner.Phone,
		Email:               partner.Email,
		Address:             partner.Address,
		City:                partner.City,
		Province:            partner.Province,
		Status:              string(partner.Status),
		StatusLabel:         partner.Status.Label(),
		SuspendedReason:     partner.SuspendedReason,
		AgreementAcceptedAt: partner.AgreementAcceptedAt,
		JoinedAt:            partner.JoinedAt,
		CreatedAt:           partner.CreatedAt,
		UpdatedAt:           partner.UpdatedAt,
	}
}

// RegisterPartnerRequest adalah body POST /partners/register dan checkout mitra baru.
type RegisterPartnerRequest struct {
	BusinessName string `json:"business_name" validate:"required,min=3,max=150"`
	PICName      string `json:"pic_name" validate:"required,min=3,max=100"`
	Phone        string `json:"phone" validate:"required,min=8,max=20"`
	Email        string `json:"email" validate:"omitempty,email"`
	Address      string `json:"address" validate:"required,max=255"`
	City         string `json:"city" validate:"required,max=100"`
	Province     string `json:"province" validate:"required,max=100"`
}

// UpdatePartnerRequest adalah body PUT /vendor/profile.
type UpdatePartnerRequest struct {
	BusinessName string `json:"business_name" validate:"required,min=3,max=150"`
	PICName      string `json:"pic_name" validate:"required,min=3,max=100"`
	Phone        string `json:"phone" validate:"required,min=8,max=20"`
	Email        string `json:"email" validate:"omitempty,email"`
	Address      string `json:"address" validate:"required,max=255"`
	City         string `json:"city" validate:"required,max=100"`
	Province     string `json:"province" validate:"required,max=100"`
}

// SuspendPartnerRequest adalah body PATCH /admin/partners/{id}/status.
type SuspendPartnerRequest struct {
	Status string `json:"status" validate:"required,oneof=active suspended expired"`
	Reason string `json:"reason" validate:"max=255"`
}
