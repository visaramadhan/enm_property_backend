package services

import (
	"context"
	"maps"
	"strings"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/httpx"
)

// UserService menangani profil pengguna dan pengelolaan akun oleh Superadmin
// melalui Firebase Admin (BE-03).
type UserService struct {
	deps Deps
}

// GetMe mengembalikan profil pengguna yang sedang masuk beserta mitranya.
func (s *UserService) GetMe(_ context.Context, user *models.User) (dto.UserView, error) {
	if user == nil {
		return dto.UserView{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}
	return s.viewWithPartner(user), nil
}

// UpdateMe memperbarui nama dan nomor telepon pengguna.
func (s *UserService) UpdateMe(ctx context.Context, user *models.User, req dto.UpdateMeRequest) (dto.UserView, error) {
	if user == nil {
		return dto.UserView{}, httpx.Unauthorized("Sesi pengguna tidak ditemukan")
	}

	user.Name = strings.TrimSpace(req.Name)
	user.Phone = normalizePhone(req.Phone)
	user.UpdatedAt = timeNow()

	if err := s.deps.Repos.Users.Update(user); err != nil {
		return dto.UserView{}, httpx.Internal(err)
	}

	_ = s.deps.Firebase.UpdateUser(ctx, user.FirebaseUID, firebase.UserParams{
		DisplayName: user.Name,
		PhoneNumber: firebasePhone(user.Phone),
	})
	return s.viewWithPartner(user), nil
}

// Get mengambil satu pengguna berdasarkan ID.
func (s *UserService) Get(id string) (dto.UserView, error) {
	user, err := s.deps.Repos.Users.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.UserView{}, httpx.NotFound("Pengguna tidak ditemukan")
		}
		return dto.UserView{}, httpx.Internal(err)
	}
	return s.viewWithPartner(user), nil
}

// List mengembalikan daftar pengguna untuk halaman manajemen akun.
func (s *UserService) List(q dto.PaginationQuery, filters map[string]string) (dto.ListResult[dto.UserView], error) {
	users, total, err := s.deps.Repos.Users.List(q.ToListQuery(filters))
	if err != nil {
		return dto.ListResult[dto.UserView]{}, httpx.Internal(err)
	}

	views := make([]dto.UserView, 0, len(users))
	for _, user := range users {
		views = append(views, s.viewWithPartner(user))
	}
	return dto.NewListResult(views, total), nil
}

// Create membuat akun baru pada Firebase Admin lalu menyalinnya ke database.
func (s *UserService) Create(ctx context.Context, req dto.AdminCreateUserRequest) (dto.UserView, error) {
	role := models.Role(req.Role)
	if !role.Valid() {
		return dto.UserView{}, httpx.BadRequest("Peran yang dipilih tidak dikenal")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := s.deps.Repos.Users.GetByEmail(email); err == nil {
		return dto.UserView{}, httpx.Conflict("Email sudah terdaftar")
	}

	uid, err := s.createFirebaseUser(ctx, email, req)
	if err != nil {
		return dto.UserView{}, err
	}

	now := timeNow()
	user := &models.User{
		ID:            newID(),
		FirebaseUID:   uid,
		Email:         email,
		Name:          strings.TrimSpace(req.Name),
		Phone:         normalizePhone(req.Phone),
		Role:          role,
		Status:        models.UserStatusActive,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.deps.Repos.Users.Create(user); err != nil {
		return dto.UserView{}, httpx.Conflict("Email atau akun sudah terdaftar")
	}

	_ = s.deps.Firebase.SetRoleClaim(ctx, uid, string(role))
	return dto.NewUserView(user), nil
}

// Update mengubah nama, nomor telepon, dan peran pengguna.
func (s *UserService) Update(ctx context.Context, id string, req dto.AdminUpdateUserRequest) (dto.UserView, error) {
	user, err := s.deps.Repos.Users.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.UserView{}, httpx.NotFound("Pengguna tidak ditemukan")
		}
		return dto.UserView{}, httpx.Internal(err)
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		user.Name = name
	}
	if req.Phone != "" {
		user.Phone = normalizePhone(req.Phone)
	}
	roleChanged := false
	if req.Role != "" {
		role := models.Role(req.Role)
		if !role.Valid() {
			return dto.UserView{}, httpx.BadRequest("Peran yang dipilih tidak dikenal")
		}
		if role != user.Role {
			user.Role = role
			roleChanged = true
		}
	}
	user.UpdatedAt = timeNow()

	if err := s.deps.Repos.Users.Update(user); err != nil {
		return dto.UserView{}, httpx.Internal(err)
	}

	_ = s.deps.Firebase.UpdateUser(ctx, user.FirebaseUID, firebase.UserParams{
		DisplayName: user.Name,
		PhoneNumber: firebasePhone(user.Phone),
	})
	if roleChanged {
		_ = s.deps.Firebase.SetRoleClaim(ctx, user.FirebaseUID, string(user.Role))
	}
	return s.viewWithPartner(user), nil
}

// SetStatus mengaktifkan atau menonaktifkan akun, termasuk pada Firebase.
// Alasan penonaktifan dicatat melalui audit log oleh handler.
func (s *UserService) SetStatus(ctx context.Context, id string, active bool, reason string) (dto.UserView, error) {
	_ = reason

	user, err := s.deps.Repos.Users.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return dto.UserView{}, httpx.NotFound("Pengguna tidak ditemukan")
		}
		return dto.UserView{}, httpx.Internal(err)
	}

	if active {
		user.Status = models.UserStatusActive
	} else {
		user.Status = models.UserStatusDisabled
	}
	user.UpdatedAt = timeNow()

	if err := s.deps.Repos.Users.Update(user); err != nil {
		return dto.UserView{}, httpx.Internal(err)
	}
	if err := s.deps.Firebase.SetDisabled(ctx, user.FirebaseUID, !active); err != nil {
		return dto.UserView{}, httpx.Unprocessable("Status akun tersimpan namun Firebase menolak perubahan: " + err.Error())
	}
	return s.viewWithPartner(user), nil
}

// ResetPassword membentuk tautan reset password Firebase untuk dikirim admin.
func (s *UserService) ResetPassword(ctx context.Context, id string) (string, error) {
	user, err := s.deps.Repos.Users.GetByID(id)
	if err != nil {
		if database.IsNotFound(err) {
			return "", httpx.NotFound("Pengguna tidak ditemukan")
		}
		return "", httpx.Internal(err)
	}
	if user.Email == "" {
		return "", httpx.Unprocessable("Pengguna tidak memiliki alamat email")
	}

	link, err := s.deps.Firebase.PasswordResetLink(ctx, user.Email)
	if err != nil {
		return "", httpx.Unprocessable("Gagal membuat tautan reset password: " + err.Error())
	}
	if link == "" {
		return "", httpx.Unprocessable("Firebase Admin belum dikonfigurasi sehingga tautan reset tidak dapat dibuat")
	}
	return link, nil
}

// createFirebaseUser membuat akun Firebase, atau UID lokal saat kredensial
// Firebase belum tersedia pada lingkungan pengembangan.
func (s *UserService) createFirebaseUser(ctx context.Context, email string, req dto.AdminCreateUserRequest) (string, error) {
	if !s.deps.Firebase.Configured() {
		if s.deps.Config.IsDevelopment() {
			return "local-" + newID(), nil
		}
		return "", httpx.Unprocessable("Firebase Admin belum dikonfigurasi sehingga akun tidak dapat dibuat")
	}

	uid, err := s.deps.Firebase.CreateUser(ctx, firebase.UserParams{
		Email:       email,
		Password:    req.Password,
		DisplayName: strings.TrimSpace(req.Name),
		PhoneNumber: firebasePhone(normalizePhone(req.Phone)),
	})
	if err != nil {
		return "", httpx.Unprocessable("Gagal membuat akun pada Firebase: " + err.Error())
	}
	return uid, nil
}

// viewWithPartner melampirkan profil mitra bila pengguna berperan mitra.
func (s *UserService) viewWithPartner(user *models.User) dto.UserView {
	view := dto.NewUserView(user)
	if partner, err := s.deps.Repos.Partners.GetByUserID(user.ID); err == nil {
		view = view.WithPartner(partner)
	}
	return view
}

// mergeFilters menyalin peta filter agar pemanggil tidak berbagi referensi.
func mergeFilters(base map[string]string, extra map[string]string) map[string]string {
	if len(extra) == 0 {
		return base
	}
	merged := make(map[string]string, len(base)+len(extra))
	maps.Copy(merged, base)
	maps.Copy(merged, extra)
	return merged
}
