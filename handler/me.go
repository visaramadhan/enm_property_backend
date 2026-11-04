package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// MeHandler mengelola profil pengguna yang sedang masuk.
type MeHandler struct {
	services *services.Services
}

// NewMeHandler membuat handler profil pengguna.
func NewMeHandler(svc *services.Services) *MeHandler {
	return &MeHandler{services: svc}
}

// Get mengembalikan profil pengguna beserta profil mitra bila ada.
func (h *MeHandler) Get(c *gin.Context) {
	view, err := h.services.User.GetMe(c.Request.Context(), middleware.CurrentUser(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Update mengubah nama dan nomor telepon pengguna.
func (h *MeHandler) Update(c *gin.Context) {
	var req dto.UpdateMeRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.User.UpdateMe(c.Request.Context(), middleware.CurrentUser(c), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}
