package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// SiteSettingHandler menyajikan pengaturan tampilan yang dipakai halaman
// publik (mis. area masuk dan daftar) tanpa autentikasi.
type SiteSettingHandler struct {
	services *services.Services
}

// NewSiteSettingHandler membuat handler pengaturan tampilan publik.
func NewSiteSettingHandler(svc *services.Services) *SiteSettingHandler {
	return &SiteSettingHandler{services: svc}
}

// Get mengembalikan nama situs, URL logo, dan URL gambar area autentikasi.
func (h *SiteSettingHandler) Get(c *gin.Context) {
	httpx.OK(c, dto.PublicSettingsView{
		SiteName:      h.services.Setting.SiteName(),
		LogoURL:       h.services.Setting.SiteLogoURL(),
		LoginImageURL: h.services.Setting.LoginImageURL(),
	})
}
