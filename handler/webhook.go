package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// MidtransWebhookHandler memproses notifikasi pembayaran Midtrans.
type MidtransWebhookHandler struct {
	services *services.Services
}

// NewMidtransWebhookHandler membuat handler webhook Midtrans.
func NewMidtransWebhookHandler(svc *services.Services) *MidtransWebhookHandler {
	return &MidtransWebhookHandler{services: svc}
}

// Notification memproses perubahan status transaksi dari Midtrans. Signature
// sudah diverifikasi middleware.MidtransSignature sebelum handler dijalankan.
func (h *MidtransWebhookHandler) Notification(c *gin.Context) {
	var payload dto.MidtransNotification
	if !middleware.BindJSON(c, &payload) {
		return
	}

	if err := h.services.Payment.HandleNotification(c.Request.Context(), payload); err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, gin.H{"received": true})
}
