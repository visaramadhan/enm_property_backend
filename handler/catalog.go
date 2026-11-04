package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// CatalogHandler melayani katalog properti publik: pencarian, detail, dan kontak.
type CatalogHandler struct {
	services *services.Services
}

// NewCatalogHandler membuat handler katalog.
func NewCatalogHandler(svc *services.Services) *CatalogHandler {
	return &CatalogHandler{services: svc}
}

// Search mengembalikan daftar properti tayang sesuai filter katalog.
func (h *CatalogHandler) Search(c *gin.Context) {
	var query dto.PropertyFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page = page
	query.PerPage = perPage

	result, err := h.services.Catalog.Search(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Detail mengembalikan detail properti tayang berdasarkan slug atau ID.
func (h *CatalogHandler) Detail(c *gin.Context) {
	var params dto.SlugParam
	if !middleware.BindURI(c, &params) {
		return
	}

	detail, err := h.services.Catalog.GetDetail(params.Slug, middleware.CurrentUser(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, detail)
}

// Contact mengembalikan nomor WhatsApp mitra atau nomor cadangan admin.
func (h *CatalogHandler) Contact(c *gin.Context) {
	var params dto.SlugParam
	if !middleware.BindURI(c, &params) {
		return
	}

	response, err := h.services.Catalog.GetContact(params.Slug, middleware.CurrentUser(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, response)
}
