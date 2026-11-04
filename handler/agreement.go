package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// AgreementHandler melayani perjanjian kemitraan yang sedang berlaku.
type AgreementHandler struct {
	services *services.Services
}

// NewAgreementHandler membuat handler perjanjian publik.
func NewAgreementHandler(svc *services.Services) *AgreementHandler {
	return &AgreementHandler{services: svc}
}

// Active mengembalikan versi perjanjian yang sedang berlaku.
func (h *AgreementHandler) Active(c *gin.Context) {
	agreement, err := h.services.Agreement.GetActive()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, agreement)
}
