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

// Charge memulai pembayaran QRIS/VA untuk satu order langganan.
func (h *CheckoutHandler) Charge(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.ChargeRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Payment.Charge(c.Request.Context(), middleware.CurrentUser(c), params.OrderCode, req.Method)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Payment mengembalikan instruksi pembayaran yang sudah dibuat untuk satu order.
func (h *CheckoutHandler) Payment(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}

	view, err := h.services.Payment.GetPayment(middleware.CurrentUser(c), params.OrderCode)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Orders mengembalikan riwayat transaksi langganan milik mitra yang sedang masuk.
func (h *CheckoutHandler) Orders(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Payment.ListOrders(middleware.CurrentUser(c), query, queryFilters(c, "status", "kind"))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}
