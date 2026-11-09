package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// queryFilters mengumpulkan filter dari query string berdasarkan daftar kunci.
func queryFilters(c *gin.Context, keys ...string) map[string]string {
	filters := make(map[string]string, len(keys))
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			filters[key] = value
		}
	}
	return filters
}

// AdminDashboardHandler menyajikan ringkasan operasional platform.
type AdminDashboardHandler struct {
	services *services.Services
}

// NewAdminDashboardHandler membuat handler dashboard admin.
func NewAdminDashboardHandler(svc *services.Services) *AdminDashboardHandler {
	return &AdminDashboardHandler{services: svc}
}

// Dashboard mengembalikan ringkasan mitra, properti, langganan, dan order.
func (h *AdminDashboardHandler) Dashboard(c *gin.Context) {
	view, err := h.services.Report.AdminDashboard()
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// AdminPartnerHandler mengelola data mitra dari sisi admin.
type AdminPartnerHandler struct {
	services *services.Services
}

// NewAdminPartnerHandler membuat handler manajemen mitra.
func NewAdminPartnerHandler(svc *services.Services) *AdminPartnerHandler {
	return &AdminPartnerHandler{services: svc}
}

// List mengembalikan daftar mitra beserta filter status dan wilayah.
func (h *AdminPartnerHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Partner.ListForAdmin(query, queryFilters(c, "status", "city", "province"))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Detail mengembalikan satu mitra berdasarkan ID.
func (h *AdminPartnerHandler) Detail(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	view, err := h.services.Partner.GetForAdmin(params.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// SetStatus mengubah status mitra menjadi aktif, ditangguhkan, atau kedaluwarsa.
func (h *AdminPartnerHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SuspendPartnerRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Partner.SetStatus(params.ID, models.PartnerStatus(req.Status), req.Reason)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForStatus(req.Status),
		Entity:   "partner",
		EntityID: params.ID,
		After:    view.Status,
		Note:     req.Reason,
	})
	httpx.OK(c, view)
}

// AdminPropertyHandler mengelola properti seluruh mitra dari sisi admin.
type AdminPropertyHandler struct {
	services *services.Services
}

// NewAdminPropertyHandler membuat handler manajemen properti.
func NewAdminPropertyHandler(svc *services.Services) *AdminPropertyHandler {
	return &AdminPropertyHandler{services: svc}
}

// List mengembalikan daftar properti seluruh mitra.
func (h *AdminPropertyHandler) List(c *gin.Context) {
	var query dto.PropertyFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Property.ListForAdmin(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// SetStatus mengubah status properti beserta alasan moderasi.
func (h *AdminPropertyHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.AdminPropertyStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.SetStatusByAdmin(params.ID, models.PropertyStatus(req.Status), req.Reason)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForPropertyStatus(req.Status),
		Entity:   "property",
		EntityID: params.ID,
		After:    detail.Status,
		Note:     req.Reason,
	})
	httpx.OK(c, detail)
}

// Create membuat properti atas nama mitra paket gratis.
func (h *AdminPropertyHandler) Create(c *gin.Context) {
	var req dto.AdminCreatePropertyRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.AdminCreateForFree(req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "property",
		EntityID: detail.ID,
		After:    detail.Title,
	})
	httpx.Created(c, detail)
}

// Update mengubah data properti atas nama mitra.
func (h *AdminPropertyHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.AdminUpdatePropertyRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	detail, err := h.services.Property.AdminUpdate(params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionUpdate,
		Entity:   "property",
		EntityID: params.ID,
		After:    detail.Title,
	})
	httpx.OK(c, detail)
}

// AdminMediaHandler mengelola foto properti atas nama mitra.
type AdminMediaHandler struct {
	services *services.Services
}

// NewAdminMediaHandler membuat handler media admin.
func NewAdminMediaHandler(svc *services.Services) *AdminMediaHandler {
	return &AdminMediaHandler{services: svc}
}

// Upload menyimpan foto pada properti milik mitra tertentu.
func (h *AdminMediaHandler) Upload(c *gin.Context) {
	var params dto.PropertyIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	property, err := h.services.Property.GetEntity(params.PropertyID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	partner, err := h.services.Partner.GetByID(property.PartnerID)
	if err != nil {
		httpx.Respond(c, err)
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

	response, err := h.services.Media.UploadPhotos(c.Request.Context(), partner, params.PropertyID, files)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, response)
}

// Delete menghapus satu foto properti milik mitra.
func (h *AdminMediaHandler) Delete(c *gin.Context) {
	var params dto.PhotoIDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	property, err := h.services.Property.GetEntity(params.PropertyID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	partner, err := h.services.Partner.GetByID(property.PartnerID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	if err := h.services.Media.DeletePhoto(c.Request.Context(), partner, params.PropertyID, params.PhotoID); err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.NoContent(c)
}

// AdminReportHandler menyajikan laporan langganan dan ekspornya.
type AdminReportHandler struct {
	services *services.Services
}

// NewAdminReportHandler membuat handler laporan admin.
func NewAdminReportHandler(svc *services.Services) *AdminReportHandler {
	return &AdminReportHandler{services: svc}
}

// Subscriptions mengembalikan laporan langganan sesuai filter.
func (h *AdminReportHandler) Subscriptions(c *gin.Context) {
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

// Detail mengembalikan satu baris laporan berdasarkan kode order.
func (h *AdminReportHandler) Detail(c *gin.Context) {
	var params dto.CodeParam
	if !middleware.BindURI(c, &params) {
		return
	}

	row, err := h.services.Report.SubscriptionDetail(params.OrderCode)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, row)
}

// Export mengunduh laporan langganan sebagai berkas XLSX atau PDF.
func (h *AdminReportHandler) Export(c *gin.Context) {
	var query dto.ReportFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	data, filename, err := exportSubscriptionReport(h.services, query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionExport,
		Entity: "subscription_report",
		Note:   filename,
	})
	download(c, data, exportContentType(query.Format), filename)
}

// AdminUserHandler mengelola akun pengguna platform.
type AdminUserHandler struct {
	services *services.Services
}

// NewAdminUserHandler membuat handler manajemen pengguna.
func NewAdminUserHandler(svc *services.Services) *AdminUserHandler {
	return &AdminUserHandler{services: svc}
}

// List mengembalikan daftar pengguna beserta filter peran dan status.
func (h *AdminUserHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.User.List(query, queryFilters(c, "role", "status"))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create membuat akun pengguna baru beserta perannya.
func (h *AdminUserHandler) Create(c *gin.Context) {
	var req dto.AdminCreateUserRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.User.Create(c.Request.Context(), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "user",
		EntityID: view.ID,
		After:    view.Email,
	})
	httpx.Created(c, view)
}

// Detail mengembalikan satu pengguna berdasarkan ID.
func (h *AdminUserHandler) Detail(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	view, err := h.services.User.Get(params.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Update mengubah nama, telepon, dan peran pengguna.
func (h *AdminUserHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.AdminUpdateUserRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.User.Update(c.Request.Context(), params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionUpdate,
		Entity:   "user",
		EntityID: params.ID,
		After:    view.Email,
	})
	httpx.OK(c, view)
}

// SetStatus menonaktifkan atau mengaktifkan kembali akun pengguna.
func (h *AdminUserHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetUserStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.User.SetStatus(c.Request.Context(), params.ID, req.Active, req.Reason)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForActive(req.Active),
		Entity:   "user",
		EntityID: params.ID,
		After:    view.Status,
		Note:     req.Reason,
	})
	httpx.OK(c, view)
}

// ResetPassword membuat tautan reset password untuk pengguna.
func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	link, err := h.services.User.ResetPassword(c.Request.Context(), params.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionResetPassword,
		Entity:   "user",
		EntityID: params.ID,
	})
	httpx.OK(c, gin.H{"reset_link": link})
}

// AdminPlanHandler mengelola paket langganan.
type AdminPlanHandler struct {
	services *services.Services
}

// NewAdminPlanHandler membuat handler paket langganan admin.
func NewAdminPlanHandler(svc *services.Services) *AdminPlanHandler {
	return &AdminPlanHandler{services: svc}
}

// List mengembalikan seluruh paket termasuk yang nonaktif.
func (h *AdminPlanHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Plan.ListAll(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create menambah paket langganan baru.
func (h *AdminPlanHandler) Create(c *gin.Context) {
	var req dto.CreatePlanRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Plan.Create(req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "plan",
		EntityID: view.ID,
		After:    view.Name,
	})
	httpx.Created(c, view)
}

// Update mengubah data paket langganan.
func (h *AdminPlanHandler) Update(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.UpdatePlanRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Plan.Update(params.ID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionUpdate,
		Entity:   "plan",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// SetStatus mengaktifkan atau menonaktifkan paket.
func (h *AdminPlanHandler) SetStatus(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}
	var req dto.SetPlanStatusRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Plan.SetActive(params.ID, req.Active)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   auditActionForActive(req.Active),
		Entity:   "plan",
		EntityID: params.ID,
		After:    view.Name,
	})
	httpx.OK(c, view)
}

// Reorder mengubah urutan tampil paket.
func (h *AdminPlanHandler) Reorder(c *gin.Context) {
	var req dto.ReorderPlansRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	views, err := h.services.Plan.Reorder(req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "plan",
		Note:   "urutan paket diperbarui",
	})
	httpx.OK(c, views)
}

// AdminAgreementHandler mengelola versi perjanjian kemitraan.
type AdminAgreementHandler struct {
	services *services.Services
}

// NewAdminAgreementHandler membuat handler perjanjian admin.
func NewAdminAgreementHandler(svc *services.Services) *AdminAgreementHandler {
	return &AdminAgreementHandler{services: svc}
}

// List mengembalikan riwayat versi perjanjian.
func (h *AdminAgreementHandler) List(c *gin.Context) {
	var query dto.PaginationQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Agreement.List(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Create menambahkan versi perjanjian baru.
func (h *AdminAgreementHandler) Create(c *gin.Context) {
	var req dto.CreateAgreementRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Agreement.CreateVersion(middleware.CurrentUser(c), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionCreate,
		Entity:   "agreement",
		EntityID: view.ID,
		After:    view.Version,
	})
	httpx.Created(c, view)
}

// Activate menjadikan satu versi perjanjian sebagai yang berlaku.
func (h *AdminAgreementHandler) Activate(c *gin.Context) {
	var params dto.IDParam
	if !middleware.BindURI(c, &params) {
		return
	}

	view, err := h.services.Agreement.Activate(params.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action:   models.AuditActionActivate,
		Entity:   "agreement",
		EntityID: params.ID,
		After:    view.Version,
	})
	httpx.OK(c, view)
}

// AdminSettingHandler membaca dan mengubah pengaturan sistem.
type AdminSettingHandler struct {
	services *services.Services
}

// NewAdminSettingHandler membuat handler pengaturan admin.
func NewAdminSettingHandler(svc *services.Services) *AdminSettingHandler {
	return &AdminSettingHandler{services: svc}
}

// Get mengembalikan seluruh pengaturan dengan nilai rahasia tersamar.
func (h *AdminSettingHandler) Get(c *gin.Context) {
	httpx.OK(c, h.services.Setting.GetAll())
}

// Update mengubah satu atau beberapa pengaturan sekaligus.
func (h *AdminSettingHandler) Update(c *gin.Context) {
	var req dto.UpdateSettingsRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	actor := middleware.CurrentUser(c)
	actorID := ""
	if actor != nil {
		actorID = actor.ID
	}

	views, err := h.services.Setting.Update(actorID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "setting",
		Note:   "pengaturan sistem diperbarui",
	})
	httpx.OK(c, views)
}

// UploadImage mengunggah gambar pengaturan (logo situs atau gambar area masuk)
// dan menyimpan URL hasilnya ke pengaturan terkait.
func (h *AdminSettingHandler) UploadImage(c *gin.Context) {
	var params dto.KindParam
	if !middleware.BindURI(c, &params) {
		return
	}

	files, err := uploadedFiles(c)
	if err != nil {
		httpx.Respond(c, httpx.BadRequest("Formulir unggahan tidak valid"))
		return
	}
	if len(files) == 0 {
		httpx.Respond(c, httpx.BadRequest("Tidak ada gambar yang diunggah"))
		return
	}

	actor := middleware.CurrentUser(c)
	actorID := ""
	if actor != nil {
		actorID = actor.ID
	}

	result, err := h.services.Setting.UploadImage(c.Request.Context(), params.Kind, files[0], actorID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "setting",
		Note:   "gambar pengaturan " + params.Kind + " diperbarui",
	})
	httpx.Created(c, result)
}

// AdminAuditHandler menyajikan dan mengekspor audit log.
type AdminAuditHandler struct {
	services *services.Services
}

// NewAdminAuditHandler membuat handler audit log.
func NewAdminAuditHandler(svc *services.Services) *AdminAuditHandler {
	return &AdminAuditHandler{services: svc}
}

// List mengembalikan daftar audit log terfilter.
func (h *AdminAuditHandler) List(c *gin.Context) {
	var query dto.AuditFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	page, perPage := pageParams(c)
	query.Page, query.PerPage = page, perPage

	result, err := h.services.Audit.List(query)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	paginate(c, result, page, perPage)
}

// Export mengunduh audit log sebagai berkas XLSX atau PDF.
func (h *AdminAuditHandler) Export(c *gin.Context) {
	var query dto.AuditFilterQuery
	if !middleware.BindQuery(c, &query) {
		return
	}

	var (
		data     []byte
		filename string
		err      error
	)
	if query.Format == "pdf" {
		data, filename, err = h.services.Audit.ExportPDF(query)
	} else {
		data, filename, err = h.services.Audit.ExportXLSX(query)
	}
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	download(c, data, exportContentType(query.Format), filename)
}

// exportSubscriptionReport memilih format berkas laporan langganan.
func exportSubscriptionReport(svc *services.Services, query dto.ReportFilterQuery) ([]byte, string, error) {
	if query.Format == "pdf" {
		return svc.Report.ExportPDF(query)
	}
	return svc.Report.ExportXLSX(query)
}

// exportContentType memetakan format ekspor menjadi tipe konten berkas.
func exportContentType(format string) string {
	if format == "pdf" {
		return "application/pdf"
	}
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

// auditActionForStatus memilih aksi audit dari status mitra baru.
func auditActionForStatus(status string) models.AuditAction {
	if status == string(models.PartnerStatusActive) {
		return models.AuditActionActivate
	}
	return models.AuditActionUpdate
}

// auditActionForPropertyStatus memilih aksi audit dari status properti baru.
func auditActionForPropertyStatus(status string) models.AuditAction {
	if status == string(models.PropertyStatusActive) {
		return models.AuditActionActivate
	}
	if status == string(models.PropertyStatusInactive) {
		return models.AuditActionDeactivate
	}
	return models.AuditActionUpdate
}

// auditActionForActive memilih aksi audit dari flag aktif.
func auditActionForActive(active bool) models.AuditAction {
	if active {
		return models.AuditActionActivate
	}
	return models.AuditActionDeactivate
}
