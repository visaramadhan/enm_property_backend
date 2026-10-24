package middleware

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"

	"enmproperty/backend/config"
	"enmproperty/backend/platform/httpx"
)

// UploadGuard membatasi total ukuran body pada route unggah foto sebelum
// multipart diurai. Validasi per berkas dilakukan ValidateUploadedFile.
func UploadGuard(cfg config.Storage) gin.HandlerFunc {
	totalLimit := cfg.MaxUploadBytes
	if cfg.MaxPhotosPerProperty > 1 {
		totalLimit = cfg.MaxUploadBytes * int64(cfg.MaxPhotosPerProperty)
	}

	return func(c *gin.Context) {
		if totalLimit > 0 && c.Request.ContentLength > totalLimit {
			httpx.Abort(c, http.StatusRequestEntityTooLarge, httpx.CodeBadRequest,
				"Total ukuran berkas yang diunggah melebihi batas")
			return
		}
		c.Next()
	}
}

// ValidateUploadedFile memeriksa ukuran, tipe dari isi berkas, dan keamanan nama berkas.
func ValidateUploadedFile(file *multipart.FileHeader, cfg config.Storage) error {
	if file == nil {
		return httpx.BadRequest("Berkas foto tidak ditemukan")
	}
	if cfg.MaxUploadBytes > 0 && file.Size > cfg.MaxUploadBytes {
		return httpx.BadRequest("Ukuran foto melebihi batas yang diizinkan")
	}
	if !safeFileName(file.Filename) {
		return httpx.BadRequest("Nama berkas tidak aman")
	}

	opened, err := file.Open()
	if err != nil {
		return httpx.BadRequest("Berkas foto tidak dapat dibaca")
	}
	defer opened.Close()

	detected, err := mimetype.DetectReader(opened)
	if err != nil {
		return httpx.BadRequest("Tipe berkas tidak dapat dikenali")
	}

	mime := strings.Split(detected.String(), ";")[0]
	if !allowedMime(mime, cfg.AllowedMimeTypes) {
		return httpx.BadRequest("Tipe berkas " + mime + " tidak diizinkan")
	}
	return nil
}

// allowedMime memeriksa apakah tipe berkas termasuk daftar yang diizinkan.
func allowedMime(mime string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if strings.EqualFold(strings.TrimSpace(candidate), mime) {
			return true
		}
	}
	return false
}

// safeFileName menolak nama berkas yang memuat pemisah path atau karakter aneh.
func safeFileName(name string) bool {
	if name == "" {
		return false
	}
	if strings.ContainsAny(name, "\x00/\\") {
		return false
	}
	base := filepath.Base(name)
	return base == name && !strings.HasPrefix(base, ".")
}
