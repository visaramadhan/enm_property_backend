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

// Get mengembalikan nama situs, judul navbar, URL logo, warna sistem, URL
// gambar area autentikasi, dan nomor kontak cadangan sistem.
func (h *SiteSettingHandler) Get(c *gin.Context) {
	httpx.OK(c, dto.PublicSettingsView{
		SiteName:        h.services.Setting.SiteName(),
		NavTitle:        h.services.Setting.SiteNavTitle(),
		LogoURL:         h.services.Setting.SiteLogoURL(),
		BrandColor:      h.services.Setting.BrandColor(),
		LoginImageURL:   h.services.Setting.LoginImageURL(),
		ContactWhatsApp: h.services.Setting.ContactFallback(),
	})
}
