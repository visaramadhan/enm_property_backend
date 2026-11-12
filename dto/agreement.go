package dto

import (
	"time"

	"enmproperty/backend/models"
)

// AgreementView adalah representasi versi perjanjian kemitraan.
type AgreementView struct {
	ID          string    `json:"id"`
	Version     string    `json:"version"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Checklist   []string  `json:"checklist"`
	IsActive    bool      `json:"is_active"`
	EffectiveAt time.Time `json:"effective_at"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewAgreementView membentuk tampilan perjanjian dari entitas domain.
func NewAgreementView(agreement *models.Agreement) AgreementView {
	if agreement == nil {
		return AgreementView{}
	}
	checklist := agreement.Checklist
	if checklist == nil {
		checklist = []string{}
	}
	return AgreementView{
		ID:          agreement.ID,
		Version:     agreement.Version,
		Title:       agreement.Title,
		Content:     agreement.Content,
		Checklist:   checklist,
		IsActive:    agreement.IsActive,
		EffectiveAt: agreement.EffectiveAt,
		CreatedBy:   agreement.CreatedBy,
		CreatedAt:   agreement.CreatedAt,
	}
}

// CreateAgreementRequest adalah body POST /admin/agreements.
type CreateAgreementRequest struct {
	Version     string     `json:"version" validate:"required,max=30"`
	Title       string     `json:"title" validate:"required,min=5,max=150"`
	Content     string     `json:"content" validate:"required,min=20"`
	Checklist   []string   `json:"checklist" validate:"max=20,dive,max=300"`
	EffectiveAt *time.Time `json:"effective_at"`
	Activate    bool       `json:"activate"`
}

// AcceptAgreementRequest adalah body POST /vendor/agreement.
type AcceptAgreementRequest struct {
	AgreementID string `json:"agreement_id"`
}

// AgreementAcceptanceView adalah ringkasan bukti persetujuan perjanjian.
type AgreementAcceptanceView struct {
	ID               string    `json:"id"`
	AgreementID      string    `json:"agreement_id"`
	AgreementVersion string    `json:"agreement_version"`
	PartnerID        string    `json:"partner_id"`
	UserID           string    `json:"user_id"`
	AcceptedAt       time.Time `json:"accepted_at"`
	IPAddress        string    `json:"ip_address"`
}
