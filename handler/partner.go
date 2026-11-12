package handler

import (
	"github.com/gin-gonic/gin"

	"enmproperty/backend/dto"
	"enmproperty/backend/middleware"
	"enmproperty/backend/platform/httpx"
	"enmproperty/backend/services"
)

// PartnerHandler melayani pendaftaran dan profil kemitraan.
type PartnerHandler struct {
	services *services.Services
}

// NewPartnerHandler membuat handler kemitraan.
func NewPartnerHandler(svc *services.Services) *PartnerHandler {
	return &PartnerHandler{services: svc}
}

// Register mendaftarkan profil mitra bagi pengguna yang sedang masuk.
func (h *PartnerHandler) Register(c *gin.Context) {
	var req dto.RegisterPartnerRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	partner, err := h.services.Partner.Register(c.Request.Context(), middleware.CurrentUser(c), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.Created(c, dto.NewPartnerView(partner))
}

// Profile mengembalikan profil mitra pengguna yang sedang masuk.
func (h *PartnerHandler) Profile(c *gin.Context) {
	view, err := h.services.Partner.GetProfile(middleware.CurrentUser(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// UpdateProfile mengubah data profil mitra.
func (h *PartnerHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdatePartnerRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	view, err := h.services.Partner.UpdateProfile(middleware.CurrentUser(c), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, view)
}

// Agreement mengembalikan perjanjian aktif beserta status persetujuan mitra
// sehingga halaman /vendor/agreement dapat menampilkan tombol setuju.
func (h *PartnerHandler) Agreement(c *gin.Context) {
	active, err := h.services.Agreement.GetActive()
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	payload := gin.H{"agreement": active, "accepted": false}
	if user := middleware.CurrentUser(c); user != nil {
		if partner, partnerErr := h.services.Partner.GetByUserID(user.ID); partnerErr == nil {
			if acceptance, acceptErr := h.services.Agreement.GetAcceptance(partner.ID, active.ID); acceptErr == nil && acceptance != nil {
				payload["accepted"] = true
				payload["acceptance"] = acceptance
			}
		}
	}
	httpx.OK(c, payload)
}

// AcceptAgreement mencatat persetujuan mitra atas versi perjanjian yang
// berlaku (mis. saat modal perjanjian pertama kali ditampilkan di dashboard).
func (h *PartnerHandler) AcceptAgreement(c *gin.Context) {
	var req dto.AcceptAgreementRequest
	if !middleware.BindJSON(c, &req) {
		return
	}

	user := middleware.CurrentUser(c)
	partner, err := h.services.Partner.GetByUserID(user.ID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	acceptance, err := h.services.Partner.AcceptAgreement(user, partner, req.AgreementID, requestMeta(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	httpx.OK(c, acceptance)
}
