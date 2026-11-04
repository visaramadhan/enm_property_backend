package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// PlanHandler melayani daftar dan detail paket langganan publik.
type PlanHandler struct {
	services *services.Services
}

// NewPlanHandler membuat handler paket langganan.
func NewPlanHandler(svc *services.Services) *PlanHandler {
	return &PlanHandler{services: svc}
}

// List mengembalikan paket aktif untuk halaman pricing.
func (h *PlanHandler) List(c *gin.Context) {
	plans, err := h.services.Plan.ListActive()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, plans)
}

// Detail mengembalikan satu paket berdasarkan slug.
func (h *PlanHandler) Detail(c *gin.Context) {
	var params dto.SlugParam
	if !middleware.BindURI(c, &params) {
		return
	}

	plan, err := h.services.Plan.GetBySlug(params.Slug)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, plan)
}
