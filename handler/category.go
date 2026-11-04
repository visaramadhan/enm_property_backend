package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// CategoryHandler melayani daftar kategori publik untuk katalog.
type CategoryHandler struct {
	services *services.Services
}

// NewCategoryHandler membuat handler kategori publik.
func NewCategoryHandler(svc *services.Services) *CategoryHandler {
	return &CategoryHandler{services: svc}
}

// List mengembalikan kategori aktif untuk filter katalog publik.
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.services.Category.ListActive()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, categories)
}

// AdminCategoryHandler mengelola master data kategori dari sisi superadmin.
type AdminCategoryHandler struct {
	services *services.Services
}

// NewAdminCategoryHandler membuat handler master data kategori.
func NewAdminCategoryHandler(svc *services.Services) *AdminCategoryHandler {
	return &AdminCategoryHandler{services: svc}
}

// List mengembalikan seluruh kategori termasuk yang nonaktif.
func (h *AdminCategoryHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Category.ListAll(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create menambah kategori baru.
func (h *AdminCategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Category.Create(req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "category",
		EntityID: view.ID,
		After:    view.Name,
	})
	httpx.Created(c, view)
}

// Update mengubah data kategori.
func (h *AdminCategoryHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.UpdateCategoryRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Category.Update(params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionUpdate,
		Entity:   "category",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// SetStatus mengaktifkan atau menonaktifkan kategori.
func (h *AdminCategoryHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetPlanStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Category.SetActive(params.ID, req.Active)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForActive(req.Active),
		Entity:   "category",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// Delete menghapus kategori dari master data.
func (h *AdminCategoryHandler) Delete(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	if err := h.services.Category.Delete(params.ID); err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionDelete,
		Entity:   "category",
		EntityID: params.ID,
	})
	httpx.NoContent(c)
}
