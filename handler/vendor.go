package handler

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// VendorDashboardHandler menyajikan ringkasan dashboard mitra.
type VendorDashboardHandler struct {
	services *services.Services
}

// NewVendorDashboardHandler membuat handler dashboard mitra.
func NewVendorDashboardHandler(svc *services.Services) *VendorDashboardHandler {
	return &VendorDashboardHandler{services: svc}
}

// Dashboard mengembalikan ringkasan properti, tayangan, dan paket mitra.
func (h *VendorDashboardHandler) Dashboard(c *gin.Context) {
	partner := middleware.CurrentPartner(c)
	if partner == nil {
		httpx.Respond(c, httpx.Forbidden("Profil mitra belum tersedia pada sesi ini"))
		return
	}

	view, err := h.services.Report.VendorDashboard(partner)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// VendorSubscriptionHandler menyajikan masa aktif paket mitra.
type VendorSubscriptionHandler struct {
	services *services.Services
}

// NewVendorSubscriptionHandler membuat handler langganan mitra.
func NewVendorSubscriptionHandler(svc *services.Services) *VendorSubscriptionHandler {
	return &VendorSubscriptionHandler{services: svc}
}

// Current mengembalikan langganan berjalan beserta sisa hari dan kuota listing.
func (h *VendorSubscriptionHandler) Current(c *gin.Context) {
	partner := middleware.CurrentPartner(c)
	if partner == nil {
		httpx.Respond(c, httpx.Forbidden("Profil mitra belum tersedia pada sesi ini"))
		return
	}

	view, err := h.services.Subscription.GetCurrent(partner)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	if view == nil {
		httpx.OK(c, dto.SubscriptionView{})
		return
	}
	httpx.OK(c, view)
}

// VendorPropertyHandler mengelola properti milik mitra.
type VendorPropertyHandler struct {
	services *services.Services
}

// NewVendorPropertyHandler membuat handler properti mitra.
func NewVendorPropertyHandler(svc *services.Services) *VendorPropertyHandler {
	return &VendorPropertyHandler{services: svc}
}

// List mengembalikan properti milik mitra beserta filter tabel.
func (h *VendorPropertyHandler) List(c *gin.Context) {
	var query dto.PropertyFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page = page
	query.PerPage = perPage

	result, err := h.services.Property.ListByPartner(middleware.CurrentPartner(c), query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create menambah properti baru untuk mitra.
func (h *VendorPropertyHandler) Create(c *gin.Context) {
	var req dto.CreatePropertyRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.Create(middleware.CurrentPartner(c), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, detail)
}

// Detail mengembalikan satu properti mitra.
func (h *VendorPropertyHandler) Detail(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	detail, err := h.services.Property.Get(params.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, detail)
}

// Update mengubah data properti mitra.
func (h *VendorPropertyHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.CreatePropertyRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.Update(middleware.CurrentPartner(c), params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, detail)
}

// Delete menghapus properti mitra.
func (h *VendorPropertyHandler) Delete(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	if err := h.services.Property.Delete(middleware.CurrentPartner(c), params.ID); err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.NoContent(c)
}

// SetStatus mengubah status tayang properti mitra.
func (h *VendorPropertyHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetPropertyStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.SetStatus(middleware.CurrentPartner(c), params.ID, models.PropertyStatus(req.Status))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, detail)
}

// SetAvailability mengubah jumlah unit tersedia properti.
func (h *VendorPropertyHandler) SetAvailability(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetAvailabilityRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.SetAvailability(middleware.CurrentPartner(c), params.ID, req.UnitsAvailable)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, detail)
}

// VendorMediaHandler mengelola foto properti milik mitra.
type VendorMediaHandler struct {
	services *services.Services
}

// NewVendorMediaHandler membuat handler media mitra.
func NewVendorMediaHandler(svc *services.Services) *VendorMediaHandler {
	return &VendorMediaHandler{services: svc}
}

// Upload menyimpan satu atau beberapa foto properti.
func (h *VendorMediaHandler) Upload(c *gin.Context) {
	var params dto.PropertyIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	files, err := uploadedFiles(c)
	if err != nil {
		httpx.Respond(c, httpx.BadRequest("Formulir unggahan tidak valid"))
		return
	}
	if len(files) == 0 {
		httpx.Respond(c, httpx.BadRequest("Tidak ada foto yang diunggah"))
		return
	}

	response, err := h.services.Media.UploadPhotos(c.Request.Context(), middleware.CurrentPartner(c), params.PropertyID, files)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, response)
}

// Delete menghapus satu foto properti.
func (h *VendorMediaHandler) Delete(c *gin.Context) {
	var params dto.PhotoIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	if err := h.services.Media.DeletePhoto(c.Request.Context(), middleware.CurrentPartner(c), params.PropertyID, params.PhotoID); err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.NoContent(c)
}

// Order mengubah urutan foto properti.
func (h *VendorMediaHandler) Order(c *gin.Context) {
	var params dto.PropertyIDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.PhotoOrderRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	photos, err := h.services.Media.Reorder(middleware.CurrentPartner(c), params.PropertyID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, photos)
}

// SetPrimary menandai satu foto sebagai foto utama properti.
func (h *VendorMediaHandler) SetPrimary(c *gin.Context) {
	var params dto.PhotoIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	photos, err := h.services.Media.SetPrimary(middleware.CurrentPartner(c), params.PropertyID, params.PhotoID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, photos)
}

// uploadedFiles mengambil berkas dari formulir multipart, menerima kunci
// "files" maupun "file" agar klien lama tetap dapat mengunggah satu foto.
func uploadedFiles(c *gin.Context) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	return files, nil
}
