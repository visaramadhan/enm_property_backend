package firebase

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"enmproperty/backend/platform/httpx"
)

// publicKeysURL adalah endpoint resmi Google yang memuat kunci publik (x509)
// untuk memverifikasi ID token Firebase.
const publicKeysURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"

// Batas masa berlaku cache kunci publik agar tidak sepenuhnya bergantung pada
// header Cache-Control dari server Google.
const (
	batasCacheMin = 5 * time.Minute
	batasCacheMax = 24 * time.Hour
)

// idTokenVerifier memverifikasi ID token Firebase memakai kunci publik Google
// sehingga masuk dengan Firebase (termasuk Google Sign-In) dapat berjalan tanpa
// service account Admin SDK.
//
// Firebase menerbitkan ID token sebagai JWT RS256 yang ditandatangani layanan
// `securetoken` Google. Kunci publiknya dipublikasikan pada [publicKeysURL] dan
// di-cache mengikuti header `Cache-Control`, sehingga verifikasi berikutnya
// tidak selalu memanggil jaringan.
//
// Verifier ini hanya untuk MEMVERIFIKASI token. Operasi pengelolaan akun
// (buat/ubah/nonaktifkan, tautan reset, custom claim) tetap membutuhkan
// kredensial Admin SDK lewat `FIREBASE_CREDENTIALS_FILE`/`_JSON`.
type idTokenVerifier struct {
	projectID string
	issuer    string
	client    *http.Client
	keysURL   string

	mu     sync.RWMutex
	keys   map[string]*rsa.PublicKey
	expiry time.Time
}

// newIDTokenVerifier membentuk verifier untuk sebuah project Firebase.
func newIDTokenVerifier(projectID string) *idTokenVerifier {
	projectID = strings.TrimSpace(projectID)
	return &idTokenVerifier{
		projectID: projectID,
		issuer:    "https://securetoken.google.com/" + projectID,
		client:    &http.Client{Timeout: 10 * time.Second},
		keysURL:   publicKeysURL,
	}
}

// verify memvalidasi tanda tangan dan klaim ID token, lalu memetakannya ke Token.
func (v *idTokenVerifier) verify(ctx context.Context, idToken string) (*Token, error) {
	claims := jwt.MapClaims{}
	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("algoritma tanda tangan %q tidak didukung", t.Method.Alg())
		}
		kid, _ := t.Header["kid"].(string)
		return v.publicKey(ctx, kid)
	}

	parsed, err := jwt.
		NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()})).
		ParseWithClaims(idToken, claims, keyFunc)
	if err != nil || !parsed.Valid {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid atau kedaluwarsa")
	}

	// Klaim wajib ID token Firebase: aud = project ID, iss = securetoken project.
	if stringClaim(claims, "aud") != v.projectID || stringClaim(claims, "iss") != v.issuer {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid atau kedaluwarsa")
	}
	uid := stringClaim(claims, "sub")
	if uid == "" {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid atau kedaluwarsa")
	}

	token := &Token{
		UID:           uid,
		Email:         stringClaim(claims, "email"),
		EmailVerified: boolClaim(claims, "email_verified"),
		Name:          stringClaim(claims, "name"),
		Picture:       stringClaim(claims, "picture"),
		PhoneNumber:   stringClaim(claims, "phone_number"),
		Role:          stringClaim(claims, "role"),
		Claims:        map[string]any(claims),
	}
	if seconds, ok := claims["iat"].(float64); ok {
		token.IssuedAt = time.Unix(int64(seconds), 0)
	}
	if seconds, ok := claims["exp"].(float64); ok {
		token.ExpiresAt = time.Unix(int64(seconds), 0)
	}
	return token, nil
}

// publicKey mengembalikan kunci publik untuk `kid` tertentu, memuat ulang cache
// bila kunci belum ada atau sudah kedaluwarsa.
func (v *idTokenVerifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	kid = strings.TrimSpace(kid)
	if kid == "" {
		return nil, httpx.Unauthorized("Token autentikasi tidak valid")
	}
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}
	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}
	if key, ok := v.cachedKey(kid); ok {
		return key, nil
	}
	return nil, httpx.Unauthorized("Token autentikasi tidak valid")
}

// cachedKey mengambil kunci dari cache bila masih berlaku.
func (v *idTokenVerifier) cachedKey(kid string) (*rsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.keys == nil || time.Now().After(v.expiry) {
		return nil, false
	}
	key, ok := v.keys[kid]
	return key, ok
}

// refreshKeys mengunduh kunci publik terbaru dari Google lalu menyimpannya.
func (v *idTokenVerifier) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.keysURL, nil)
	if err != nil {
		return httpx.Internal(err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return httpx.Internal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return httpx.Internal(fmt.Errorf("gagal mengambil kunci publik Firebase: HTTP %d", resp.StatusCode))
	}

	var certs map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return httpx.Internal(err)
	}

	keys := make(map[string]*rsa.PublicKey, len(certs))
	for kid, isi := range certs {
		block, _ := pem.Decode([]byte(isi))
		if block == nil {
			continue
		}
		sertifikat, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		if kunci, ok := sertifikat.PublicKey.(*rsa.PublicKey); ok {
			keys[kid] = kunci
		}
	}
	if len(keys) == 0 {
		return httpx.Internal(fmt.Errorf("kunci publik Firebase tidak dapat dibaca"))
	}

	v.mu.Lock()
	v.keys = keys
	v.expiry = time.Now().Add(cacheDuration(resp.Header.Get("Cache-Control")))
	v.mu.Unlock()
	return nil
}

// cacheDuration membaca `max-age` dari header Cache-Control dengan batas aman.
func cacheDuration(cacheControl string) time.Duration {
	durasi := batasCacheMin
	for _, bagian := range strings.Split(cacheControl, ",") {
		bagian = strings.TrimSpace(bagian)
		if !strings.HasPrefix(bagian, "max-age=") {
			continue
		}
		if detik, err := strconv.Atoi(strings.TrimPrefix(bagian, "max-age=")); err == nil {
			durasi = time.Duration(detik) * time.Second
		}
	}
	if durasi < batasCacheMin {
		return batasCacheMin
	}
	if durasi > batasCacheMax {
		return batasCacheMax
	}
	return durasi
}
