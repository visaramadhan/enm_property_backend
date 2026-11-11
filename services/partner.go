package services

import (
	"context"
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// PartnerService menangani registrasi dan pengelolaan profil kemitraan (BE-04).
type PartnerService struct {
	deps          Deps
	agreements    *AgreementService
	subscriptions *SubscriptionService
}

// Register membuat profil kemitraan baru yang langsung aktif tanpa verifikasi
// admin. Paket Gratis otomatis diaktifkan agar mitra dapat segera memasang
// listing.
func (s *PartnerService) Register(ctx context.Context, user *models.User, req dto.RegisterPartnerRequest) (*models.Partner, error) {
	if user == nil {
		return nil, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}
	if existing, err := s.deps.Repos.Partners.GetByUserID(user.ID); err == nil && existing != nil {
		return nil, httpx.Conflict("Akun Anda sudah terdaftar sebagai mitra")
	}

	now := timeNow()
	partner := &models.Partner{
		ID:        newID(),
		UserID:    user.ID,
		Status:    models.PartnerStatusActive,
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.applyProfile(partner, req.BusinessName, req.PICName, req.Phone, req.Email, req.Address, req.City, req.Province)

	if err := s.deps.Repos.Partners.Create(partner); err != nil {
		return nil, httpx.Internal(err)
	}
	s.ensureFreePlan(partner.ID)
	if err := s.promoteToMitra(ctx, user); err != nil {
		return nil, err
	}
	return partner, nil
}

// Ensure mengembalikan profil mitra pengguna, membuat atau memperbaruinya dari
// data yang dikirim pada alur checkout.
func (s *PartnerService) Ensure(ctx context.Context, user *models.User, req dto.RegisterPartnerRequest) (*models.Partner, error) {
	if user == nil {
		return nil, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	partner, err := s.deps.Repos.Partners.GetByUserID(user.ID)
	if err != nil {
		if !database.IsNotFound(err) {
			return nil, httpx.Internal(err)
		}
		partner = nil
	}

	now := timeNow()
	if partner == nil {
		partner = &models.Partner{
			ID:        newID(),
			UserID:    user.ID,
			Status:    models.PartnerStatusActive,
			JoinedAt:  &now,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.applyProfile(partner, req.BusinessName, req.PICName, req.Phone, req.Email, req.Address, req.City, req.Province)
		if err := s.deps.Repos.Partners.Create(partner); err != nil {
			return nil, httpx.Internal(err)
		}
		s.ensureFreePlan(partner.ID)
	} else {
		s.applyProfile(partner, req.BusinessName, req.PICName, req.Phone, req.Email, req.Address, req.City, req.Province)
		partner.UpdatedAt = now
		if err := s.deps.Repos.Partners.Update(partner); err != nil {
			return nil, httpx.Internal(err)
		}
	}

	if err := s.promoteToMitra(ctx, user); err != nil {
		return nil, err
	}
	return partner, nil
}

// EnsureFromUser memastikan pengguna memiliki profil kemitraan, membuatnya dari
// data akun bila belum ada.
//
// Dipakai alur checkout paket pada kemitraan baru: profil awal cukup memuat data
// akun (nama, email, telepon) agar order dapat dibuat, sedangkan data badan usaha
// dilengkapi mitra setelah pembayaran berhasil (BE-04/BE-06).
func (s *PartnerService) EnsureFromUser(ctx context.Context, user *models.User) (*models.Partner, error) {
	if user == nil {
		return nil, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	partner, err := s.deps.Repos.Partners.GetByUserID(user.ID)
	if err == nil && partner != nil {
		return partner, nil
	}
	if err != nil && !database.IsNotFound(err) {
		return nil, httpx.Internal(err)
	}

	nama := strings.TrimSpace(user.DisplayName())
	if nama == "" {
		if indeks := strings.Index(user.Email, "@"); indeks > 0 {
			nama = user.Email[:indeks]
		}
	}
	if nama == "" {
		nama = "Mitra"
	}

	now := timeNow()
	partner = &models.Partner{
		ID:        newID(),
		UserID:    user.ID,
		Status:    models.PartnerStatusActive,
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.applyProfile(partner, nama, nama, user.Phone, user.Email, "", "", "")

	if err := s.deps.Repos.Partners.Create(partner); err != nil {
		return nil, httpx.Internal(err)
	}
	s.ensureFreePlan(partner.ID)
	if err := s.promoteToMitra(ctx, user); err != nil {
		return nil, err
	}
	return partner, nil
}

// AcceptAgreement mencatat persetujuan perjanjian atas nama mitra (BE-06).
func (s *PartnerService) AcceptAgreement(user *models.User, partner *models.Partner, agreementID string, meta RequestMeta) (*dto.AgreementAcceptanceView, error) {
	return s.agreements.Accept(user, partner, agreementID, meta)
}

// GetProfile mengembalikan profil mitra pengguna yang sedang masuk.
func (s *PartnerService) GetProfile(user *models.User) (dto.PartnerView, error) {
	partner, err := s.partnerOf(user)
	if err != nil {
		return dto.PartnerView{}, err
	}
	return dto.NewPartnerView(partner), nil
}

// UpdateProfile memperbarui profil mitra pengguna yang sedang masuk.
func (s *PartnerService) UpdateProfile(user *models.User, req dto.UpdatePartnerRequest) (dto.PartnerView, error) {
	partner, err := s.partnerOf(user)
	if err != nil {
		return dto.PartnerView{}, err
	}

	s.applyProfile(partner, req.BusinessName, req.PICName, req.Phone, req.Email, req.Address, req.City, req.Province)
	partner.UpdatedAt = timeNow()
	if err := s.deps.Repos.Partners.Update(partner); err != nil {
		return dto.PartnerView{}, httpx.Internal(err)
	}
	return dto.NewPartnerView(partner), nil
}

// GetByID mengambil mitra berdasarkan ID.
func (s *PartnerService) GetByID(id string) (*models.Partner, error) {
	partner, err := s.deps.Repos.Partners.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Mitra tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	return partner, nil
}

// GetByUserID mengambil mitra berdasarkan ID pengguna.
func (s *PartnerService) GetByUserID(userID string) (*models.Partner, error) {
	return s.deps.Repos.Partners.GetByUserID(userID)
}

// Activate mengaktifkan kemitraan setelah pembayaran atau verifikasi admin.
func (s *PartnerService) Activate(id string) (dto.PartnerView, error) {
	return s.SetStatus(id, models.PartnerStatusActive, "")
}

// Suspend menonaktifkan kemitraan beserta alasannya.
func (s *PartnerService) Suspend(id, reason string) (dto.PartnerView, error) {
	return s.SetStatus(id, models.PartnerStatusSuspended, reason)
}

// Unsuspend mengaktifkan kembali kemitraan yang dinonaktifkan.
func (s *PartnerService) Unsuspend(id string) (dto.PartnerView, error) {
	return s.SetStatus(id, models.PartnerStatusActive, "")
}

// SetStatus mengubah status kemitraan secara umum.
func (s *PartnerService) SetStatus(id string, status models.PartnerStatus, reason string) (dto.PartnerView, error) {
	partner, err := s.GetByID(id)
	if err != nil {
		return dto.PartnerView{}, err
	}

	now := timeNow()
	partner.Status = status
	partner.UpdatedAt = now
	if status == models.PartnerStatusActive {
		partner.SuspendedReason = ""
		if partner.JoinedAt == nil {
			partner.JoinedAt = &now
		}
	} else {
		partner.SuspendedReason = strings.TrimSpace(reason)
	}

	if err := s.deps.Repos.Partners.Update(partner); err != nil {
		return dto.PartnerView{}, httpx.Internal(err)
	}
	return dto.NewPartnerView(partner), nil
}

// ListForAdmin mengembalikan daftar mitra untuk halaman manajemen.
func (s *PartnerService) ListForAdmin(q dto.PaginationQuery, filters map[string]string) (dto.ListResult[dto.PartnerView], error) {
	partners, total, err := s.deps.Repos.Partners.List(q.ToListQuery(filters))
	if err != nil {
		return dto.ListResult[dto.PartnerView]{}, httpx.Internal(err)
	}

	views := make([]dto.PartnerView, 0, len(partners))
	for _, partner := range partners {
		views = append(views, dto.NewPartnerView(partner))
	}
	return dto.NewListResult(views, total), nil
}

// GetForAdmin mengambil detail satu mitra untuk halaman admin.
func (s *PartnerService) GetForAdmin(id string) (dto.PartnerView, error) {
	partner, err := s.GetByID(id)
	if err != nil {
		return dto.PartnerView{}, err
	}
	return dto.NewPartnerView(partner), nil
}

// partnerOf mengambil profil mitra milik pengguna, atau galat bila belum ada.
func (s *PartnerService) partnerOf(user *models.User) (*models.Partner, error) {
	if user == nil {
		return nil, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}
	partner, err := s.deps.Repos.Partners.GetByUserID(user.ID)
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Profil kemitraan belum dibuat")
		}
		return nil, httpx.Internal(err)
	}
	return partner, nil
}

// applyProfile menyalin data permintaan ke entitas mitra.
func (s *PartnerService) applyProfile(partner *models.Partner, businessName, picName, phone, email, address, city, province string) {
	partner.BusinessName = strings.TrimSpace(businessName)
	partner.PICName = strings.TrimSpace(picName)
	partner.Phone = normalizePhone(phone)
	partner.Email = strings.ToLower(strings.TrimSpace(email))
	partner.Address = strings.TrimSpace(address)
	partner.City = strings.TrimSpace(city)
	partner.Province = strings.TrimSpace(province)
}

// ensureFreePlan mengaktifkan paket Gratis untuk mitra baru agar dapat langsung
// memasang listing tanpa verifikasi admin maupun pembayaran.
func (s *PartnerService) ensureFreePlan(partnerID string) {
	if s.subscriptions != nil {
		s.subscriptions.EnsureFree(partnerID)
	}
}

// promoteToMitra memastikan pengguna berperan mitra beserta custom claim Firebase.
func (s *PartnerService) promoteToMitra(ctx context.Context, user *models.User) error {
	if user.Role == models.RoleMitra {
		return nil
	}
	user.Role = models.RoleMitra
	user.UpdatedAt = timeNow()
	if err := s.deps.Repos.Users.Update(user); err != nil {
		return httpx.Internal(err)
	}
	_ = s.deps.Firebase.SetRoleClaim(ctx, user.FirebaseUID, string(models.RoleMitra))
	return nil
}
