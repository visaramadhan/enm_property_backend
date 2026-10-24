package middleware

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/httpx"
)

// EnsureUserOwns memastikan sumber daya milik pengguna yang sedang masuk.
// Peran pengelola platform (admin, superadmin, owner) diperbolehkan melewati.
func EnsureUserOwns(c *gin.Context, ownerUserID, resource string) error {
	user := CurrentUser(c)
	if user == nil {
		return httpx.Unauthorized("Autentikasi diperlukan")
	}
	if user.Role.IsInternal() || user.ID == ownerUserID {
		return nil
	}
	return httpx.Forbidden("Anda tidak berhak mengakses " + resource + " ini")
}

// EnsurePartnerOwns memastikan sumber daya milik mitra pengguna yang sedang masuk.
func EnsurePartnerOwns(c *gin.Context, partnerID, resource string) error {
	if role := CurrentRole(c); role.IsInternal() {
		return nil
	}

	partner := CurrentPartner(c)
	if partner == nil {
		return httpx.Forbidden("Profil mitra belum tersedia pada sesi ini")
	}
	if partner.ID != partnerID {
		return httpx.Forbidden("Anda tidak berhak mengakses " + resource + " ini")
	}
	return nil
}
