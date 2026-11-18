package services

import (
	"context"
	"log/slog"
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// AccountService menangani siklus hidup akun pengguna, termasuk penghapusan
// akun beserta seluruh data terkait (right to erasure).
type AccountService struct {
	deps Deps
}

// DeleteAccount menghapus akun pengguna yang sedang masuk beserta seluruh data
// terkait: profil mitra, properti (termasuk yang sudah dihapus lunak), foto,
// favorit, langganan, order, dan bukti persetujuan. Setelah data DB terhapus
// (dalam satu transaksi), akun Firebase serta berkas foto fisik dibersihkan
// secara best-effort agar kegagalan pihak ketiga tidak membatalkan permintaan.
func (s *AccountService) DeleteAccount(ctx context.Context, user *models.User, meta RequestMeta) error {
	if user == nil {
		return httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	partnerID := ""
	propertyIDs := make([]string, 0)
	storagePaths := make([]string, 0)

	partner, err := s.deps.Repos.Partners.GetByUserID(user.ID)
	if err == nil && partner != nil {
		partnerID = partner.ID
		properties, err := s.deps.Repos.Properties.ListAllByPartner(partnerID)
		if err != nil {
			return httpx.Internal(err)
		}
		for _, property := range properties {
			propertyIDs = append(propertyIDs, property.ID)
			for _, photo := range property.Photos {
				if path := strings.TrimSpace(photo.StoragePath); path != "" {
					storagePaths = append(storagePaths, path)
				}
			}
		}
	} else if err != nil && !database.IsNotFound(err) {
		return httpx.Internal(err)
	}

	if err := s.deps.Repos.Users.DeleteUserCascade(ctx, user.ID, partnerID, propertyIDs); err != nil {
		return httpx.Internal(err)
	}

	// Firebase: kegagalan di sini tidak membatalkan penghapusan data karena
	// transaksi DB sudah commit; cukup dicatat untuk ditindaklanjuti manual.
	if err := s.deps.Firebase.DeleteUser(ctx, user.FirebaseUID); err != nil {
		slog.WarnContext(ctx, "gagal menghapus akun Firebase saat penghapusan akun",
			"user_id", user.ID, "cause", err)
	}

	// Berkas foto fisik dihapus secara best-effort.
	for _, path := range storagePaths {
		if err := s.deps.Storage.Delete(ctx, path); err != nil {
			slog.WarnContext(ctx, "gagal menghapus berkas foto saat penghapusan akun",
				"user_id", user.ID, "path", path, "cause", err)
		}
	}

	// Audit log tanpa data pribadi (nama/email) agar alamat email dapat dipakai
	// kembali untuk pendaftaran akun baru.
	entry := &models.AuditLog{
		ID:        newID(),
		Action:    models.AuditActionDelete,
		Entity:    "user",
		EntityID:  user.ID,
		Note:      "Penghapusan akun oleh pengguna",
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: timeNow(),
	}
	if err := s.deps.Repos.Audits.Create(entry); err != nil {
		slog.WarnContext(ctx, "gagal mencatat audit penghapusan akun",
			"user_id", user.ID, "cause", err)
	}

	return nil
}
