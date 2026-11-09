package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// LandingHandler menyajikan konten landing page untuk halaman publik dan
// mengelolanya dari CMS (admin, superadmin, serta owner mode baca).
type LandingHandler struct {
	services *services.Services
}

// NewLandingHandler membuat handler konten landing page.
func NewLandingHandler(svc *services.Services) *LandingHandler {
	return &LandingHandler{services: svc}
}

// Get mengembalikan konten landing page. Dipakai endpoint publik maupun CMS.
func (h *LandingHandler) Get(c *gin.Context) {
	httpx.OK(c, h.services.Landing.Content())
}

// Update menyimpan perubahan konten landing page dari CMS.
func (h *LandingHandler) Update(c *gin.Context) {
	var req dto.UpdateLandingRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	actor := middleware.CurrentUser(c)
	actorID := ""
	if actor != nil {
		actorID = actor.ID
	}

	view, err := h.services.Landing.Update(actorID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "landing",
		Note:   "konten landing page diperbarui",
	})
	httpx.OK(c, view)
}

// Publish memicu revalidasi ISR halaman landing page pada frontend.
func (h *LandingHandler) Publish(c *gin.Context) {
	if err := h.services.Landing.Revalidate(c.Request.Context()); err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "landing",
		Note:   "landing page dipublikasikan (revalidasi ISR)",
	})
	httpx.OK(c, dto.LandingPublishResponse{
		Message:     "Landing page berhasil dipublikasikan",
		Revalidated: true,
	})
}

// UploadImage menyimpan gambar landing (hero/og) setelah dikonversi ke WebP.
func (h *LandingHandler) UploadImage(c *gin.Context) {
	var params dto.LandingImageParam
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

	result, err := h.services.Landing.UploadImage(c.Request.Context(), params.Kind, files[0])
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	middleware.SetAudit(c, middleware.AuditEntry{
		Action: models.AuditActionUpdate,
		Entity: "landing",
		Note:   "gambar landing diperbarui (" + params.Kind + ")",
	})
	httpx.Created(c, result)
}
