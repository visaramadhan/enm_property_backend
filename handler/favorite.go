package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// FavoriteHandler mengelola daftar favorit customer.
type FavoriteHandler struct {
	services *services.Services
}

// NewFavoriteHandler membuat handler favorit.
func NewFavoriteHandler(svc *services.Services) *FavoriteHandler {
	return &FavoriteHandler{services: svc}
}

// List mengembalikan properti favorit customer yang sedang masuk.
func (h *FavoriteHandler) List(c *gin.Context) {
	result, err := h.services.Favorite.List(middleware.CurrentUser(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	page, perPage := pageParams(c)
	paginate(c, result, page, perPage)
}

// Add menambahkan properti ke daftar favorit.
func (h *FavoriteHandler) Add(c *gin.Context) {
	var params dto.PropertyIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	response, err := h.services.Favorite.Add(middleware.CurrentUser(c), params.PropertyID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, response)
}

// Remove menghapus properti dari daftar favorit.
func (h *FavoriteHandler) Remove(c *gin.Context) {
	var params dto.PropertyIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	response, err := h.services.Favorite.Remove(middleware.CurrentUser(c), params.PropertyID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, response)
}
