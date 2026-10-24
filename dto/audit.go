package dto

import (
	"time"

	"enmproperty/backend/models"
)

// AuditView adalah representasi satu catatan audit.
type AuditView struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actor_id"`
	ActorName  string    `json:"actor_name"`
	ActorEmail string    `json:"actor_email"`
	ActorRole  string    `json:"actor_role"`
	Action     string    `json:"action"`
	Entity     string    `json:"entity"`
	EntityID   string    `json:"entity_id"`
	Before     string    `json:"before,omitempty"`
	After      string    `json:"after,omitempty"`
	Note       string    `json:"note,omitempty"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewAuditView membentuk tampilan audit dari entitas domain.
func NewAuditView(entry *models.AuditLog) AuditView {
	if entry == nil {
		return AuditView{}
	}
	return AuditView{
		ID:         entry.ID,
		ActorID:    entry.ActorID,
		ActorName:  entry.ActorName,
		ActorEmail: entry.ActorEmail,
		ActorRole:  string(entry.ActorRole),
		Action:     string(entry.Action),
		Entity:     entry.Entity,
		EntityID:   entry.EntityID,
		Before:     entry.Before,
		After:      entry.After,
		Note:       entry.Note,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  entry.CreatedAt,
	}
}

// NewAuditViews membentuk daftar tampilan audit.
func NewAuditViews(entries []*models.AuditLog) []AuditView {
	views := make([]AuditView, 0, len(entries))
	for _, entry := range entries {
		views = append(views, NewAuditView(entry))
	}
	return views
}

// AuditFilterQuery adalah parameter daftar dan ekspor audit log.
type AuditFilterQuery struct {
	Page      int    `form:"page"`
	PerPage   int    `form:"per_page"`
	Search    string `form:"search"`
	Entity    string `form:"entity" validate:"omitempty,max=60"`
	EntityID  string `form:"entity_id" validate:"omitempty,max=64"`
	ActorID   string `form:"actor_id" validate:"omitempty,max=64"`
	ActorRole string `form:"actor_role" validate:"omitempty,oneof=customer mitra admin superadmin owner"`
	Action    string `form:"action" validate:"omitempty,oneof=create update delete activate deactivate reset_password export login"`
	From      string `form:"from" validate:"omitempty"`
	To        string `form:"to" validate:"omitempty"`
	Format    string `form:"format" validate:"omitempty,oneof=xlsx pdf"`
}
