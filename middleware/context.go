// Package middleware berisi seluruh middleware HTTP sesuai urutan eksekusi pada
// Bab 4.5 dokumen spesifikasi, mulai dari recovery hingga penanganan error.
package middleware

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/models"
	"enmproperty/backend/platform/firebase"
)

// Kunci konteks yang dipakai bersama antar middleware dan handler.
const (
	contextTokenKey        = "auth_token"
	contextUserKey         = "auth_user"
	contextPartnerKey      = "auth_partner"
	contextSubscriptionKey = "auth_subscription"
	contextAuditKey        = "audit_entry"
)

// Token mengambil klaim ID token yang sudah diverifikasi.
func Token(c *gin.Context) *firebase.Token {
	if value, exists := c.Get(contextTokenKey); exists {
		if token, ok := value.(*firebase.Token); ok {
			return token
		}
	}
	return nil
}

// CurrentUser mengambil pengguna terautentikasi dari konteks.
func CurrentUser(c *gin.Context) *models.User {
	if value, exists := c.Get(contextUserKey); exists {
		if user, ok := value.(*models.User); ok {
			return user
		}
	}
	return nil
}

// SetUser menyimpan pengguna terautentikasi pada konteks.
func SetUser(c *gin.Context, user *models.User) {
	c.Set(contextUserKey, user)
}

// CurrentPartner mengambil profil mitra aktif dari konteks.
func CurrentPartner(c *gin.Context) *models.Partner {
	if value, exists := c.Get(contextPartnerKey); exists {
		if partner, ok := value.(*models.Partner); ok {
			return partner
		}
	}
	return nil
}

// SetPartner menyimpan profil mitra pada konteks.
func SetPartner(c *gin.Context, partner *models.Partner) {
	c.Set(contextPartnerKey, partner)
}

// CurrentRole mengembalikan peran pengguna, atau string kosong bila belum login.
func CurrentRole(c *gin.Context) models.Role {
	if user := CurrentUser(c); user != nil {
		return user.Role
	}
	if token := Token(c); token != nil {
		return models.Role(token.Role)
	}
	return ""
}
