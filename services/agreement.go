package services

import (
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// AgreementService mengelola versi perjanjian kemitraan dan bukti persetujuannya
// (BE-06).
type AgreementService struct {
	deps Deps
}

// GetActive mengembalikan versi perjanjian yang sedang berlaku.
func (s *AgreementService) GetActive() (dto.AgreementView, error) {
	agreement, err := s.deps.Repos.Agreements.GetActive()
	if err != nil {
		if database.IsNotFound(err) {
			return dto.AgreementView{}, httpx.NotFound("Belum ada perjanjian yang berlaku")
		}
		return dto.AgreementView{}, httpx.Internal(err)
	}
	return dto.NewAgreementView(agreement), nil
}

// List mengembalikan riwayat versi perjanjian untuk halaman admin.
func (s *AgreementService) List(q dto.PaginationQuery) (dto.ListResult[dto.AgreementView], error) {
	agreements, total, err := s.deps.Repos.Agreements.List(q.ToListQuery(nil))
	if err != nil {
		return dto.ListResult[dto.AgreementView]{}, httpx.Internal(err)
	}
	views := make([]dto.AgreementView, 0, len(agreements))
	for _, agreement := range agreements {
		views = append(views, dto.NewAgreementView(agreement))
	}
	return dto.NewListResult(views, total), nil
}

// Get mengambil satu versi perjanjian berdasarkan ID.
func (s *AgreementService) Get(id string) (dto.AgreementView, error) {
	agreement, err := s.deps.Repos.Agreements.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.AgreementView{}, httpx.NotFound("Perjanjian tidak ditemukan")
		}
		return dto.AgreementView{}, httpx.Internal(err)
	}
	return dto.NewAgreementView(agreement), nil
}

// CreateVersion menambah versi perjanjian baru dan opsional langsung mengaktifkannya.
func (s *AgreementService) CreateVersion(actor *models.User, req dto.CreateAgreementRequest) (dto.AgreementView, error) {
	now := timeNow()
	effectiveAt := now
	if req.EffectiveAt != nil {
		effectiveAt = *req.EffectiveAt
	}

	agreement := &models.Agreement{
		ID:          newID(),
		Version:     strings.TrimSpace(req.Version),
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		IsActive:    req.Activate,
		EffectiveAt: effectiveAt,
		CreatedAt:   now,
	}
	if actor != nil {
		agreement.CreatedBy = actor.ID
	}

	if req.Activate {
		if err := s.deactivateOthers(agreement.ID); err != nil {
			return dto.AgreementView{}, err
		}
	}
	if err := s.deps.Repos.Agreements.Create(agreement); err != nil {
		return dto.AgreementView{}, httpx.Internal(err)
	}
	return dto.NewAgreementView(agreement), nil
}

// Activate menjadikan satu versi perjanjian sebagai versi yang berlaku.
func (s *AgreementService) Activate(id string) (dto.AgreementView, error) {
	agreement, err := s.deps.Repos.Agreements.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.AgreementView{}, httpx.NotFound("Perjanjian tidak ditemukan")
		}
		return dto.AgreementView{}, httpx.Internal(err)
	}
	if err := s.deactivateOthers(id); err != nil {
		return dto.AgreementView{}, err
	}
	agreement.IsActive = true
	if err := s.deps.Repos.Agreements.Update(agreement); err != nil {
		return dto.AgreementView{}, httpx.Internal(err)
	}
	return dto.NewAgreementView(agreement), nil
}

// Accept menyimpan bukti persetujuan mitra pada versi perjanjian tertentu.
// Pemanggilan berulang bersifat idempoten.
func (s *AgreementService) Accept(user *models.User, partner *models.Partner, agreementID string, meta RequestMeta) (*dto.AgreementAcceptanceView, error) {
	if partner == nil {
		return nil, httpx.BadRequest("Profil kemitraan belum tersedia")
	}

	agreement, err := s.resolve(agreementID)
	if err != nil {
		return nil, err
	}

	if existing, err := s.deps.Repos.Agreements.GetAcceptance(partner.ID, agreement.ID); err == nil {
		view := acceptanceView(existing)
		return &view, nil
	}

	now := timeNow()
	acceptance := &models.AgreementAcceptance{
		ID:               newID(),
		AgreementID:      agreement.ID,
		AgreementVersion: agreement.Version,
		PartnerID:        partner.ID,
		AcceptedAt:       now,
		IPAddress:        meta.IPAddress,
		UserAgent:        meta.UserAgent,
	}
	if user != nil {
		acceptance.UserID = user.ID
	}
	if err := s.deps.Repos.Agreements.SaveAcceptance(acceptance); err != nil {
		return nil, httpx.Internal(err)
	}

	if partner.AgreementAcceptedAt == nil || partner.AgreementAcceptedAt.Before(now) {
		partner.AgreementAcceptedAt = &now
		partner.UpdatedAt = now
		if err := s.deps.Repos.Partners.Update(partner); err != nil {
			return nil, httpx.Internal(err)
		}
	}

	view := acceptanceView(acceptance)
	return &view, nil
}

// GetAcceptance mengembalikan bukti persetujuan mitra pada satu versi perjanjian.
func (s *AgreementService) GetAcceptance(partnerID, agreementID string) (*dto.AgreementAcceptanceView, error) {
	agreement, err := s.resolve(agreementID)
	if err != nil {
		return nil, err
	}
	acceptance, err := s.deps.Repos.Agreements.GetAcceptance(partnerID, agreement.ID)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Bukti persetujuan belum ada")
		}
		return nil, httpx.Internal(err)
	}
	view := acceptanceView(acceptance)
	return &view, nil
}

// resolve mengambil perjanjian berdasarkan ID, atau perjanjian aktif bila ID kosong.
func (s *AgreementService) resolve(agreementID string) (*models.Agreement, error) {
	if strings.TrimSpace(agreementID) == "" {
		agreement, err := s.deps.Repos.Agreements.GetActive()
		if err != nil {
			if database.IsNotFound(err) {
				return nil, httpx.Unprocessable("Belum ada perjanjian yang berlaku")
			}
			return nil, httpx.Internal(err)
		}
		return agreement, nil
	}

	agreement, err := s.deps.Repos.Agreements.GetByID(agreementID)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Perjanjian tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return agreement, nil
}

// deactivateOthers menonaktifkan seluruh versi perjanjian selain id tertentu.
func (s *AgreementService) deactivateOthers(exceptID string) error {
	agreements, _, err := s.deps.Repos.Agreements.List(database.ListQuery{PerPage: 0})
	if err != nil {
		return httpx.Internal(err)
	}
	for _, agreement := range agreements {
		if agreement.ID == exceptID || !agreement.IsActive {
			continue
		}
		agreement.IsActive = false
		if err := s.deps.Repos.Agreements.Update(agreement); err != nil {
			return httpx.Internal(err)
		}
	}
	return nil
}

// acceptanceView membentuk tampilan bukti persetujuan.
func acceptanceView(acceptance *models.AgreementAcceptance) dto.AgreementAcceptanceView {
	if acceptance == nil {
		return dto.AgreementAcceptanceView{}
	}
	return dto.AgreementAcceptanceView{
		ID:               acceptance.ID,
		AgreementID:      acceptance.AgreementID,
		AgreementVersion: acceptance.AgreementVersion,
		PartnerID:        acceptance.PartnerID,
		UserID:           acceptance.UserID,
		AcceptedAt:       acceptance.AcceptedAt,
		IPAddress:        acceptance.IPAddress,
	}
}
