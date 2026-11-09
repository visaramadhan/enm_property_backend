package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"enmproperty/backend/database"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/firebase"
	"enmproperty/backend/platform/httpx"
)

const bearerPrefix = "Bearer "

// Authenticate memverifikasi ID token Firebase dan menyimpan klaimnya pada konteks.
// Dipakai oleh endpoint yang perlu token valid namun belum tentu punya baris pengguna,
// misalnya POST /auth/session.
func Authenticate(client *firebase.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !verifyToken(c, client) {
			return
		}
		c.Next()
	}
}

// LoadUser memuat baris pengguna berdasarkan klaim token dan menolak akun nonaktif.
func LoadUser(users UserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !attachUser(c, users) {
			return
		}
		c.Next()
	}
}

// FirebaseAuth menggabungkan verifikasi token dan pemuatan pengguna aktif dalam satu
// middleware sehingga handler selalu menerima konteks pengguna yang lengkap.
func FirebaseAuth(client *firebase.Client, users UserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !verifyToken(c, client) || !attachUser(c, users) {
			return
		}
		c.Next()
	}
}

// OptionalAuth mencoba memverifikasi token bila header Authorization ada, namun
// tidak pernah menggagalkan permintaan. Dipakai endpoint publik yang tetap ingin
// mencatat pengguna bila pengunjung sedang masuk (mis. form kontak properti):
// pengunjung anonim tetap dilayani, sedangkan pengunjung yang login tercatat.
func OptionalAuth(client *firebase.Client, users UserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearerToken(c)
		if raw == "" {
			c.Next()
			return
		}
		token, err := client.VerifyIDToken(c.Request.Context(), raw)
		if err != nil {
			c.Next()
			return
		}
		if user, err := users.GetByFirebaseUID(token.UID); err == nil && user.IsActive() {
			SetUser(c, user)
		}
		c.Next()
	}
}

// verifyToken memvalidasi header Authorization dan menaruh klaim token pada konteks.
// Mengembalikan false bila permintaan sudah digagalkan dan c.Next() tidak boleh dipanggil.
func verifyToken(c *gin.Context, client *firebase.Client) bool {
	raw := bearerToken(c)
	if raw == "" {
		httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Header Authorization tidak ditemukan")
		return false
	}

	token, err := client.VerifyIDToken(c.Request.Context(), raw)
	if err != nil {
		httpx.Respond(c, err)
		return false
	}

	c.Set(contextTokenKey, token)
	return true
}

// attachUser memuat baris pengguna dari klaim token yang sudah diverifikasi dan
// menaruh pengguna aktif pada konteks.
func attachUser(c *gin.Context, users UserReader) bool {
	token := Token(c)
	if token == nil {
		httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Sesi tidak ditemukan")
		return false
	}

	user, err := users.GetByFirebaseUID(token.UID)
	if err != nil {
		if database.IsNotFound(err) {
			httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Sesi pengguna tidak ditemukan, silakan masuk kembali")
			return false
		}
		httpx.Respond(c, httpx.Internal(err))
		return false
	}
	if !user.IsActive() {
		httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Akun Anda sedang dinonaktifkan")
		return false
	}

	SetUser(c, user)
	return true
}

// RequireRole membatasi akses berdasarkan peran. Pengguna berperan owner
// selalu dibatasi hanya untuk metode GET.
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role := CurrentRole(c)
		if role == "" {
			httpx.Abort(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "Autentikasi diperlukan")
			return
		}
		if _, ok := allowed[role]; !ok {
			httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Anda tidak memiliki akses ke sumber daya ini")
			return
		}
		if role == models.RoleOwner && c.Request.Method != http.MethodGet {
			httpx.Abort(c, http.StatusForbidden, httpx.CodeForbidden, "Peran owner hanya dapat membaca data")
			return
		}

		c.Next()
	}
}

// bearerToken mengambil token dari header Authorization berformat Bearer.
func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	if len(header) > len(bearerPrefix) && strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(header[len(bearerPrefix):])
	}
	return strings.TrimSpace(header)
}
