package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// CheckoutHandler melayani pembelian paket dan pemantauan status transaksi.
type CheckoutHandler struct {
	services *services.Services
}

// NewCheckoutHandler membuat handler checkout langganan.
func NewCheckoutHandler(svc *services.Services) *CheckoutHandler {
	return &CheckoutHandler{services: svc}
}

// Checkout membuat order langganan baru beserta token pembayaran.
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var req dto.CheckoutRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	response, err := h.services.Payment.Checkout(c.Request.Context(), middleware.CurrentUser(c), req, requestMeta(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, response)
}

// Status mengembalikan status terkini satu order beserta langganannya.
func (h *CheckoutHandler) Status(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}

	view, err := h.services.Payment.GetStatus(middleware.CurrentUser(c), params.OrderCode)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}
