package services

import (
	"context"
	"strings"
	"time"

	"enmproperty/backend/database"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/httpx"
)

// AuthService menangani sinkronisasi pengguna Firebase dengan tabel pengguna
// aplikasi serta pengaturan custom claim role (BE-02).
type AuthService struct {
	deps Deps
}

// SyncSession membuat atau memperbarui baris pengguna dari klaim ID token,
// lalu menyamakan custom claim role dengan role yang tersimpan.
func (s *AuthService) SyncSession(ctx context.Context, token *firebase.Token, meta RequestMeta) (*models.User, error) {
	if token == nil || strings.TrimSpace(token.UID) == "" {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid")
	}

	user, err := s.deps.Repos.Users.GetByFirebaseUID(token.UID)
	if err != nil && !database.IsNotFound(err) {
		return nil, httpx.Internal(err)
	}

	now := timeNow()
	if user == nil {
		created, createErr := s.createFromToken(ctx, token, now)
		if createErr != nil {
			return nil, createErr
		}
		user = created
	} else {
		if err := s.refreshFromToken(ctx, user, token, now); err != nil {
			return nil, err
		}
	}

	if !user.IsActive() {
		return nil, httpx.Forbidden("Akun Anda sedang dinonaktifkan")
	}

	// Custom claim disamakan dengan role pada database agar token berikutnya
	// membawa peran yang benar.
	if err := s.deps.Firebase.SetRoleClaim(ctx, user.FirebaseUID, string(user.Role)); err != nil {
		_ = err
	}
	_ = meta
	return user, nil
}

// VerifyToken memverifikasi ID token Firebase tanpa menyentuh basis data.
func (s *AuthService) VerifyToken(ctx context.Context, idToken string) (*firebase.Token, error) {
	return s.deps.Firebase.VerifyIDToken(ctx, idToken)
}

// SetRoleClaim menulis custom claim role pada akun Firebase.
func (s *AuthService) SetRoleClaim(ctx context.Context, uid string, role models.Role) error {
	return s.deps.Firebase.SetRoleClaim(ctx, uid, string(role))
}

// createFromToken membuat baris pengguna baru dari klaim token.
func (s *AuthService) createFromToken(ctx context.Context, token *firebase.Token, now time.Time) (*models.User, error) {
	role := models.Role(token.Role)
	if !role.Valid() {
		role = models.RoleCustomer
	}
	email := strings.ToLower(strings.TrimSpace(token.Email))
	name := strings.TrimSpace(token.Name)
	if name == "" {
		name = displayNameFromEmail(email)
	}

	user := &models.User{
		ID:            newID(),
		FirebaseUID:   token.UID,
		Email:         email,
		Name:          name,
		Phone:         normalizePhone(token.PhoneNumber),
		AvatarURL:     token.Picture,
		Role:          role,
		Status:        models.UserStatusActive,
		EmailVerified: token.EmailVerified,
		LastLoginAt:   &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.deps.Repos.Users.Create(user); err != nil {
		if _, lookupErr := s.deps.Repos.Users.GetByEmail(email); lookupErr == nil {
			return nil, httpx.Conflict("Email sudah terdaftar dengan metode masuk lain")
		}
		return nil, httpx.Internal(err)
	}
	return user, nil
}

// refreshFromToken memperbarui profil pengguna dari klaim token terbaru.
func (s *AuthService) refreshFromToken(ctx context.Context, user *models.User, token *firebase.Token, now time.Time) error {
	if email := strings.ToLower(strings.TrimSpace(token.Email)); email != "" {
		user.Email = email
	}
	if name := strings.TrimSpace(token.Name); name != "" {
		user.Name = name
	}
	if user.Name == "" {
		user.Name = displayNameFromEmail(user.Email)
	}
	if token.Picture != "" {
		user.AvatarURL = token.Picture
	}
	if user.Phone == "" && token.PhoneNumber != "" {
		user.Phone = normalizePhone(token.PhoneNumber)
	}
	if token.EmailVerified {
		user.EmailVerified = true
	}
	user.LastLoginAt = &now
	user.UpdatedAt = now

	if err := s.deps.Repos.Users.Update(user); err != nil {
		return httpx.Internal(err)
	}
	return nil
}
