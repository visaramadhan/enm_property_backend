package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// OwnerHandler menyajikan dashboard dan laporan tingkat pemilik platform.
// Peran owner bersifat hanya-baca (GET), ditegakkan melalui middleware.RequireRole.
type OwnerHandler struct {
	services *services.Services
}

// NewOwnerHandler membuat handler dashboard dan laporan owner.
func NewOwnerHandler(svc *services.Services) *OwnerHandler {
	return &OwnerHandler{services: svc}
}

// Dashboard mengembalikan ringkasan platform beserta mitra teratas.
func (h *OwnerHandler) Dashboard(c *gin.Context) {
	view, err := h.services.Report.OwnerDashboard()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Subscriptions mengembalikan laporan langganan sesuai filter.
func (h *OwnerHandler) Subscriptions(c *gin.Context) {
	var query dto.ReportFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	report, err := h.services.Report.SubscriptionReport(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, report)
}

// Partners mengembalikan laporan mitra beserta propertinya.
func (h *OwnerHandler) Partners(c *gin.Context) {
	var query dto.ReportFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	if query.Page <= 0 {
		query.Page = page
	}
	if query.PerPage <= 0 {
		query.PerPage = perPage
	}

	result, err := h.services.Report.OwnerPartnerReport(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, query.Page, query.PerPage)
}
