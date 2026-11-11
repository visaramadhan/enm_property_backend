package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// FacilityHandler melayani daftar fasilitas publik untuk pilihan mitra.
type FacilityHandler struct {
	services *services.Services
}

// NewFacilityHandler membuat handler fasilitas publik.
func NewFacilityHandler(svc *services.Services) *FacilityHandler {
	return &FacilityHandler{services: svc}
}

// List mengembalikan fasilitas aktif untuk pilihan mitra.
func (h *FacilityHandler) List(c *gin.Context) {
	facilities, err := h.services.Facility.ListActive()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, facilities)
}

// AdminFacilityHandler mengelola master data fasilitas dari sisi superadmin.
type AdminFacilityHandler struct {
	services *services.Services
}

// NewAdminFacilityHandler membuat handler master data fasilitas.
func NewAdminFacilityHandler(svc *services.Services) *AdminFacilityHandler {
	return &AdminFacilityHandler{services: svc}
}

// List mengembalikan seluruh fasilitas termasuk yang nonaktif.
func (h *AdminFacilityHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Facility.ListAll(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create menambah fasilitas baru.
func (h *AdminFacilityHandler) Create(c *gin.Context) {
	var req dto.CreateFacilityRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Facility.Create(req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "facility",
		EntityID: view.ID,
		After:    view.Name,
	})
	httpx.Created(c, view)
}

// Update mengubah data fasilitas.
func (h *AdminFacilityHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.UpdateFacilityRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Facility.Update(params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionUpdate,
		Entity:   "facility",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// SetStatus mengaktifkan atau menonaktifkan fasilitas.
func (h *AdminFacilityHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetPlanStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Facility.SetActive(params.ID, req.Active)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForActive(req.Active),
		Entity:   "facility",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// Delete menghapus fasilitas dari master data.
func (h *AdminFacilityHandler) Delete(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	if err := h.services.Facility.Delete(params.ID); err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionDelete,
		Entity:   "facility",
		EntityID: params.ID,
	})
	httpx.NoContent(c)
}
