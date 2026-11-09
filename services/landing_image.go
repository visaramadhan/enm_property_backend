package services

import (
	"bytes"
	"context"
	"mime/multipart"
	"path"
	"strings"

	"github.com/deepteams/webp"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"

	"enmproperty/backend/config"
	"enmproperty/backend/dto"
	"enmproperty/backend/platform/httpx"
)

// Batasan gambar landing page sesuai spesifikasi CMS:
// konversi paksa ke WebP lossy 80%, lebar maksimum 1920px, ukuran < 250KB.
const (
	landingHeroWidth    = 1920
	landingOGWidth      = 1200
	landingMaxBytes     = 250 * 1024
	landingQuality      = 80
	landingMinQuality   = 50
	landingQualityStep  = 10
	landingWebPMimeType = "image/webp"
	landingHeroKind     = "hero"
	landingOGKind       = "og"
)

// UploadImage mengubah gambar unggahan menjadi WebP lalu menyimpannya pada
// storage. `kind` bernilai "hero" (lebar maks 1920px) atau "og" (lebar 1200px).
func (s *LandingService) UploadImage(ctx context.Context, kind string, file *multipart.FileHeader) (dto.LandingImageResponse, error) {
	kind = strings.TrimSpace(kind)
	if kind != landingHeroKind && kind != landingOGKind {
		return dto.LandingImageResponse{}, httpx.BadRequest("Jenis gambar landing tidak dikenal")
	}
	if err := s.validateImageFile(file); err != nil {
		return dto.LandingImageResponse{}, err
	}

	raw, err := readUpload(file)
	if err != nil {
		return dto.LandingImageResponse{}, err
	}

	maxWidth := landingHeroWidth
	if kind == landingOGKind {
		maxWidth = landingOGWidth
	}

	processed, width, height, err := processLandingImage(raw, maxWidth)
	if err != nil {
		return dto.LandingImageResponse{}, err
	}

	objectPath := path.Join("landing", kind+"-"+newID()+".webp")
	url, err := s.deps.Storage.Save(ctx, objectPath, processed)
	if err != nil {
		return dto.LandingImageResponse{}, httpx.Internal(err)
	}

	return dto.LandingImageResponse{
		URL:       url,
		Kind:      kind,
		MimeType:  landingWebPMimeType,
		Width:     width,
		Height:    height,
		SizeBytes: int64(len(processed)),
	}, nil
}

// validateImageFile memeriksa ukuran dan tipe berkas gambar landing.
func (s *LandingService) validateImageFile(file *multipart.FileHeader) error {
	return validateImageUpload(s.deps.Config.Storage, file)
}

// validateImageUpload memeriksa ukuran dan tipe berkas gambar terhadap
// konfigurasi storage. Dipakai bersama oleh unggahan landing dan pengaturan.
func validateImageUpload(cfg config.Storage, file *multipart.FileHeader) error {
	if file == nil {
		return httpx.BadRequest("Berkas gambar tidak ditemukan")
	}
	if cfg.MaxUploadBytes > 0 && file.Size > cfg.MaxUploadBytes {
		return httpx.BadRequest("Ukuran gambar melebihi batas yang diizinkan")
	}

	opened, err := file.Open()
	if err != nil {
		return httpx.BadRequest("Berkas gambar tidak dapat dibaca")
	}
	defer opened.Close()

	detected, err := mimetype.DetectReader(opened)
	if err != nil {
		return httpx.BadRequest("Tipe berkas tidak dapat dikenali")
	}
	mime := strings.Split(detected.String(), ";")[0]
	if !allowedMimeType(mime, cfg.AllowedMimeTypes) {
		return httpx.BadRequest("Tipe berkas " + mime + " tidak diizinkan")
	}
	return nil
}

// processLandingImage mengubah gambar menjadi WebP lossy dan menurunkan mutu
// secara bertahap sampai ukurannya berada di bawah batas yang ditetapkan.
func processLandingImage(raw []byte, maxWidth int) ([]byte, int, int, error) {
	img, err := imaging.Decode(bytes.NewReader(raw), imaging.AutoOrientation(true))
	if err != nil {
		return nil, 0, 0, httpx.BadRequest("Berkas gambar tidak dapat diproses")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, 0, 0, httpx.BadRequest("Ukuran gambar tidak valid")
	}
	if maxWidth > 0 && bounds.Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
		bounds = img.Bounds()
	}
	width, height := bounds.Dx(), bounds.Dy()

	var output []byte
	for quality := landingQuality; quality >= landingMinQuality; quality -= landingQualityStep {
		buffer := new(bytes.Buffer)
		options := &webp.EncoderOptions{
			Lossless: false,
			Quality:  float32(quality),
			Method:   4,
			Preset:   webp.PresetPhoto,
		}
		if err := webp.Encode(buffer, img, options); err != nil {
			return nil, 0, 0, httpx.Internal(err)
		}
		output = buffer.Bytes()
		if len(output) <= landingMaxBytes {
			break
		}
	}

	return output, width, height, nil
}
