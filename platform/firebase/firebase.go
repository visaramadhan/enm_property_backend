// Package firebase membungkus Firebase Admin SDK untuk verifikasi ID token,
// pengaturan custom claim role, dan pengelolaan akun dari sisi Superadmin.
//
// Saat AUTH_DEV_MODE aktif di luar produksi, satu token uji diterima tanpa
// memanggil Firebase sehingga alur pengembangan dapat berjalan tanpa kredensial.
package firebase

import (
	"context"
	"strings"
	"time"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"enmproperty/backend/config"
	"enmproperty/backend/platform/httpx"
)

// Token adalah klaim ID token Firebase yang dipakai lapisan autentikasi.
type Token struct {
	UID           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	PhoneNumber   string
	Role          string
	Claims        map[string]any
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

// Client membungkus klien Auth Firebase beserta mode pengembangan.
type Client struct {
	auth     *auth.Client
	devMode  bool
	devToken string
}

// New membuat klien Firebase Admin. Bila kredensial belum tersedia, klien tetap
// dibuat dalam mode terbatas dan hanya token uji yang akan diterima.
func New(ctx context.Context, cfg config.Firebase, authCfg config.Auth) (*Client, error) {
	client := &Client{
		devMode:  authCfg.DevMode,
		devToken: authCfg.DevToken,
	}

	if cfg.CredentialsFile == "" && cfg.CredentialsJSON == "" {
		return client, nil
	}

	var opts []option.ClientOption
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	} else {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	}

	app, err := fb.NewApp(ctx, &fb.Config{
		ProjectID:     cfg.ProjectID,
		StorageBucket: cfg.StorageBucket,
	}, opts...)
	if err != nil {
		return nil, err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	client.auth = authClient
	return client, nil
}

// DevMode melaporkan apakah token uji diterima.
func (c *Client) DevMode() bool { return c.devMode }

// Configured melaporkan apakah kredensial Firebase Admin sudah tersedia sehingga
// operasi pengelolaan akun (buat, ubah, nonaktifkan, tautan reset) dapat jalan.
func (c *Client) Configured() bool { return c.auth != nil }

// VerifyIDToken memverifikasi ID token dan mengembalikan klaim yang dipakai sistem.
func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*Token, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, httpx.Unauthorized("Token autentikasi tidak ditemukan")
	}

	if c.devMode && (idToken == c.devToken || strings.HasPrefix(idToken, c.devToken+":")) {
		return c.devTokenClaims(idToken), nil
	}

	if c.auth == nil {
		return nil, httpx.Unauthorized("Layanan autentikasi belum dikonfigurasi")
	}

	decoded, err := c.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid atau kedaluwarsa")
	}

	token := &Token{
		UID:           decoded.UID,
		Email:         stringClaim(decoded.Claims, "email"),
		EmailVerified: boolClaim(decoded.Claims, "email_verified"),
		Name:          stringClaim(decoded.Claims, "name"),
		Picture:       stringClaim(decoded.Claims, "picture"),
		PhoneNumber:   stringClaim(decoded.Claims, "phone_number"),
		Role:          stringClaim(decoded.Claims, "role"),
		Claims:        decoded.Claims,
	}
	if seconds, ok := decoded.Claims["iat"].(float64); ok {
		token.IssuedAt = time.Unix(int64(seconds), 0)
	}
	if seconds, ok := decoded.Claims["exp"].(float64); ok {
		token.ExpiresAt = time.Unix(int64(seconds), 0)
	}
	return token, nil
}

// SetRoleClaim menulis custom claim role pada akun Firebase.
func (c *Client) SetRoleClaim(ctx context.Context, uid, role string) error {
	if c.auth == nil {
		return nil
	}
	return c.auth.SetCustomUserClaims(ctx, uid, map[string]any{"role": role})
}

// UserParams adalah data pembuatan akun baru di Firebase.
type UserParams struct {
	Email       string
	Password    string
	DisplayName string
	PhoneNumber string
}

// CreateUser membuat akun Firebase dan mengembalikan UID-nya.
func (c *Client) CreateUser(ctx context.Context, params UserParams) (string, error) {
	if c.auth == nil {
		return "", httpx.Unprocessable("Firebase Admin belum dikonfigurasi sehingga akun tidak dapat dibuat")
	}
	toCreate := (&auth.UserToCreate{}).
		Email(params.Email).
		Password(params.Password).
		DisplayName(params.DisplayName)
	if params.PhoneNumber != "" {
		toCreate = toCreate.PhoneNumber(params.PhoneNumber)
	}
	record, err := c.auth.CreateUser(ctx, toCreate)
	if err != nil {
		return "", err
	}
	return record.UID, nil
}

// UpdateUser memperbarui profil akun Firebase.
func (c *Client) UpdateUser(ctx context.Context, uid string, params UserParams) error {
	if c.auth == nil {
		return nil
	}
	toUpdate := (&auth.UserToUpdate{})
	if params.Email != "" {
		toUpdate = toUpdate.Email(params.Email)
	}
	if params.DisplayName != "" {
		toUpdate = toUpdate.DisplayName(params.DisplayName)
	}
	if params.PhoneNumber != "" {
		toUpdate = toUpdate.PhoneNumber(params.PhoneNumber)
	}
	if params.Password != "" {
		toUpdate = toUpdate.Password(params.Password)
	}
	_, err := c.auth.UpdateUser(ctx, uid, toUpdate)
	return err
}

// SetDisabled menonaktifkan atau mengaktifkan akun Firebase.
func (c *Client) SetDisabled(ctx context.Context, uid string, disabled bool) error {
	if c.auth == nil {
		return nil
	}
	_, err := c.auth.UpdateUser(ctx, uid, (&auth.UserToUpdate{}).Disabled(disabled))
	return err
}

// PasswordResetLink membuat tautan reset password untuk dikirim admin ke pengguna.
func (c *Client) PasswordResetLink(ctx context.Context, email string) (string, error) {
	if c.auth == nil {
		return "", nil
	}
	return c.auth.PasswordResetLink(ctx, email)
}

// devTokenClaims membentuk klaim dari token uji berformat "<token>" atau
// "<token>:<firebaseUID>". Peran ditebak dari UID agar pengujian per role mudah.
func (c *Client) devTokenClaims(idToken string) *Token {
	uid := "dev-superadmin"
	if parts := strings.SplitN(idToken, ":", 2); len(parts) == 2 && parts[1] != "" {
		uid = parts[1]
	}

	role := "superadmin"
	email := uid + "@enmproperty.local"
	switch {
	case strings.Contains(uid, "mitra"):
		role = "mitra"
	case strings.Contains(uid, "superadmin"):
		role = "superadmin"
	case strings.Contains(uid, "admin"):
		role = "admin"
	case strings.Contains(uid, "owner"):
		role = "owner"
	case strings.Contains(uid, "customer"):
		role = "customer"
	}

	now := time.Now()
	return &Token{
		UID:           uid,
		Email:         email,
		EmailVerified: true,
		Name:          "Pengguna Uji",
		Role:          role,
		Claims:        map[string]any{"role": role, "email": email},
		IssuedAt:      now,
		ExpiresAt:     now.Add(time.Hour),
	}
}

func stringClaim(claims map[string]any, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func boolClaim(claims map[string]any, key string) bool {
	if value, ok := claims[key].(bool); ok {
		return value
	}
	return false
}
