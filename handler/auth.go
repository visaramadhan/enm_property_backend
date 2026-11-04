package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// AuthHandler menyinkronkan sesi Firebase dengan baris pengguna lokal.
type AuthHandler struct {
	services *services.Services
}

// NewAuthHandler membuat handler autentikasi.
func NewAuthHandler(svc *services.Services) *AuthHandler {
	return &AuthHandler{services: svc}
}

// Session membuat atau memperbarui pengguna lokal dari ID token Firebase yang
// sudah diverifikasi middleware.Authenticate.
func (h *AuthHandler) Session(c *gin.Context) {
	token := middleware.Token(c)
	if token == nil {
		httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Token autentikasi tidak ditemukan")
		return
	}

	user, err := h.services.Auth.SyncSession(c.Request.Context(), token, requestMeta(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	view := dto.NewUserView(user)
	if user.Role == models.RoleMitra {
		if partner, partnerErr := h.services.Partner.GetByUserID(user.ID); partnerErr == nil {
			view = view.WithPartner(partner)
		}
	}

	httpx.OK(c, dto.SessionResponse{User: view, Role: string(user.Role)})
}
