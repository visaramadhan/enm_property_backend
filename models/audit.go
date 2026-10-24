package models

import "time"

// AuditLog adalah catatan aksi admin dan superadmin pada modul BE-15.
type AuditLog struct {
	ID         string      `json:"id"`
	ActorID    string      `json:"actor_id"`
	ActorName  string      `json:"actor_name"`
	ActorEmail string      `json:"actor_email"`
	ActorRole  Role        `json:"actor_role"`
	Action     AuditAction `json:"action"`
	Entity     string      `json:"entity"`
	EntityID   string      `json:"entity_id"`
	Before     string      `json:"before,omitempty"`
	After      string      `json:"after,omitempty"`
	Note       string      `json:"note,omitempty"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
	CreatedAt  time.Time   `json:"created_at"`
}
