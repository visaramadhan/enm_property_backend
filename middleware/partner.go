package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/database"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// RequireActivePartner memastikan profil mitra ada dan berstatus aktif sebelum
// mengizinkan penulisan pada endpoint /vendor/*.
func RequireActivePartner(partners PartnerReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Autentikasi diperlukan")
			return
		}

		partner, err := partners.GetByUserID(user.ID)
		if err != nil {
			if database.IsNotFound(err) {
				httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Profil mitra belum terdaftar")
				return
			}
			httpx.Respond(c, httpx.Internal(err))
			return
		}
		if !partner.IsActive() {
			httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden,
				"Status mitra Anda "+partner.Status.Label()+", hubungi admin untuk mengaktifkan kembali")
			return
		}

		SetPartner(c, partner)
		c.Next()
	}
}

// RequireActivePackage menolak penulisan properti bila masa aktif paket mitra
// sudah berakhir. Pemeriksaan kuota listing dilakukan di lapisan service.
func RequireActivePackage(subscriptions SubscriptionReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		partner := CurrentPartner(c)
		if partner == nil {
			httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Profil mitra belum tersedia pada sesi ini")
			return
		}

		subscription, err := subscriptions.GetActiveByPartner(partner.ID)
		if err != nil {
			if database.IsNotFound(err) {
				httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Anda belum memiliki paket langganan aktif")
				return
			}
			httpx.Respond(c, httpx.Internal(err))
			return
		}
		if !subscription.IsActiveAt(time.Now()) {
			httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden,
				"Masa aktif paket Anda sudah berakhir, silakan perpanjang terlebih dahulu")
			return
		}

		c.Set(contextSubscriptionKey, subscription)
		c.Next()
	}
}

// CurrentSubscription mengambil langganan aktif yang ditempelkan middleware.
func CurrentSubscription(c *gin.Context) *models.Subscription {
	if value, exists := c.Get(contextSubscriptionKey); exists {
		if subscription, ok := value.(*models.Subscription); ok {
			return subscription
		}
	}
	return nil
}
