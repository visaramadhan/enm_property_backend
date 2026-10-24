package middleware

import (
	"enmproperty/backend/models"
)

// UserReader adalah kebutuhan minimal middleware autentikasi terhadap data pengguna.
type UserReader interface {
	GetByFirebaseUID(uid string) (*models.User, error)
}

// PartnerReader adalah kebutuhan minimal middleware kemitraan.
type PartnerReader interface {
	GetByUserID(userID string) (*models.Partner, error)
}

// SubscriptionReader adalah kebutuhan minimal pemeriksaan masa aktif paket.
type SubscriptionReader interface {
	GetActiveByPartner(partnerID string) (*models.Subscription, error)
}

// AuditWriter adalah kebutuhan minimal pencatatan audit log.
type AuditWriter interface {
	Create(entry *models.AuditLog) error
}
