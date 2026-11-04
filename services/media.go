package services

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"

	"enmproperty/backend/database"
	"enmproperty/backend/dto"
	"enmproperty/backend/models"
	"enmproperty/backend/platform/httpx"
)

// maxPhotoWidth adalah lebar maksimum foto properti setelah diubah ukurannya.
// Foto yang lebih kecil dibiarkan apa adanya agar tidak kehilangan ketajaman.
const maxPhotoWidth = 1600

// MediaService mengelola unggahan foto properti: validasi, perubahan ukuran,
// penyimpanan, pengurutan, dan penandaan foto utama (BE-10).
type MediaService struct {
	deps Deps
}

// EnforcePhotoLimit memastikan jumlah foto setelah penambahan tidak melewati
// batas STORAGE_MAX_PHOTOS_PER_PROPERTY.
func (s *MediaService) EnforcePhotoLimit(propertyID string, incoming int) error {
	limit := s.deps.Config.Storage.MaxPhotosPerProperty
	if limit <= 0 {
		return nil
	}
	photos, err := s.deps.Repos.Properties.ListPhotos(propertyID)
	if err != nil {
		return httpx.Internal(err)
	}
	if len(photos)+incoming > limit {
		return httpx.Unprocessable(fmt.Sprintf("Jumlah foto melebihi batas %d foto per properti", limit))
	}
	return nil
}

// UploadPhoto memvalidasi, mengubah ukuran, lalu menyimpan satu foto properti.
func (s *MediaService) UploadPhoto(ctx context.Context, partner *models.Partner, propertyID string, file *multipart.FileHeader) (dto.PhotoView, error) {
	property, err := s.ownedProperty(partner, propertyID)
	if err != nil {
		return dto.PhotoView{}, err
	}
	if err := s.EnforcePhotoLimit(property.ID, 1); err != nil {
		return dto.PhotoView{}, err
	}
	if err := s.validateFile(file); err != nil {
		return dto.PhotoView{}, err
	}

	raw, err := readUpload(file)
	if err != nil {
		return dto.PhotoView{}, err
	}
	processed, width, height, err := processImage(raw)
	if err != nil {
		return dto.PhotoView{}, err
	}

	existing, err := s.deps.Repos.Properties.ListPhotos(property.ID)
	if err != nil {
		return dto.PhotoView{}, httpx.Internal(err)
	}

	photoID := newID()
	objectPath := path.Join("properties", property.ID, photoID+".jpg")
	url, err := s.deps.Storage.Save(ctx, objectPath, processed)
	if err != nil {
		return dto.PhotoView{}, httpx.Internal(err)
	}

	photo := &models.PropertyPhoto{
		ID:          photoID,
		PropertyID:  property.ID,
		URL:         url,
		StoragePath: objectPath,
		MimeType:    "image/jpeg",
		SizeBytes:   int64(len(processed)),
		Width:       width,
		Height:      height,
		SortOrder:   len(existing),
		IsPrimary:   len(existing) == 0,
		CreatedAt:   timeNow(),
	}
	if err := s.deps.Repos.Properties.CreatePhoto(photo); err != nil {
		return dto.PhotoView{}, httpx.Internal(err)
	}
	return dto.NewPhotoView(photo), nil
}

// UploadPhotos menyimpan beberapa foto sekaligus dan mengembalikan ringkasannya.
func (s *MediaService) UploadPhotos(ctx context.Context, partner *models.Partner, propertyID string, files []*multipart.FileHeader) (dto.UploadPhotoResponse, error) {
	property, err := s.ownedProperty(partner, propertyID)
	if err != nil {
		return dto.UploadPhotoResponse{}, err
	}
	if len(files) == 0 {
		return dto.UploadPhotoResponse{}, httpx.BadRequest("Tidak ada foto yang diunggah")
	}
	if err := s.EnforcePhotoLimit(property.ID, len(files)); err != nil {
		return dto.UploadPhotoResponse{}, err
	}

	uploaded := make([]dto.PhotoView, 0, len(files))
	for _, file := range files {
		view, err := s.UploadPhoto(ctx, partner, property.ID, file)
		if err != nil {
			return dto.UploadPhotoResponse{}, err
		}
		uploaded = append(uploaded, view)
	}

	total, err := s.deps.Repos.Properties.ListPhotos(property.ID)
	if err != nil {
		return dto.UploadPhotoResponse{}, httpx.Internal(err)
	}
	return dto.UploadPhotoResponse{
		PropertyID: property.ID,
		Uploaded:   uploaded,
		Total:      len(total),
		Limit:      s.deps.Config.Storage.MaxPhotosPerProperty,
	}, nil
}

// DeletePhoto menghapus satu foto properti beserta berkasnya pada storage.
func (s *MediaService) DeletePhoto(ctx context.Context, partner *models.Partner, propertyID, photoID string) error {
	property, err := s.ownedProperty(partner, propertyID)
	if err != nil {
		return err
	}

	photo, err := s.photoOf(property.ID, photoID)
	if err != nil {
		return err
	}
	if photo.StoragePath != "" {
		_ = s.deps.Storage.Delete(ctx, photo.StoragePath)
	}
	if err := s.deps.Repos.Properties.DeletePhoto(photo.ID); err != nil {
		if database.IsNotFound(err) {
			return httpx.NotFound("Foto tidak ditemukan")
		}
		return httpx.Internal(err)
	}

	// Pastikan selalu ada foto utama setelah penghapusan.
	if photo.IsPrimary {
		s.promoteFirst(property.ID)
	}
	return nil
}

// Reorder mengubah urutan foto properti sesuai daftar ID yang dikirim.
func (s *MediaService) Reorder(partner *models.Partner, propertyID string, req dto.PhotoOrderRequest) ([]dto.PhotoView, error) {
	property, err := s.ownedProperty(partner, propertyID)
	if err != nil {
		return nil, err
	}

	for index, id := range req.PhotoIDs {
		photo, err := s.photoOf(property.ID, id)
		if err != nil {
			return nil, err
		}
		photo.SortOrder = index
		photo.IsPrimary = index == 0
		if err := s.deps.Repos.Properties.UpdatePhoto(photo); err != nil {
			return nil, httpx.Internal(err)
		}
	}
	return s.listViews(property.ID)
}

// SetPrimary menandai satu foto sebagai foto utama properti.
func (s *MediaService) SetPrimary(partner *models.Partner, propertyID, photoID string) ([]dto.PhotoView, error) {
	property, err := s.ownedProperty(partner, propertyID)
	if err != nil {
		return nil, err
	}

	photos, err := s.deps.Repos.Properties.ListPhotos(property.ID)
	if err != nil {
		return nil, httpx.Internal(err)
	}

	found := false
	for _, photo := range photos {
		primary := photo.ID == photoID
		if primary {
			found = true
		}
		if photo.IsPrimary == primary {
			continue
		}
		photo.IsPrimary = primary
		if err := s.deps.Repos.Properties.UpdatePhoto(photo); err != nil {
			return nil, httpx.Internal(err)
		}
	}
	if !found {
		return nil, httpx.NotFound("Foto tidak ditemukan")
	}
	return s.listViews(property.ID)
}

// ownedProperty mengambil properti dan memastikan kepemilikannya oleh mitra.
func (s *MediaService) ownedProperty(partner *models.Partner, propertyID string) (*models.Property, error) {
	if partner == nil {
		return nil, httpx.BadRequest("Profil kemitraan tidak tersedia")
	}
	property, err := s.deps.Repos.Properties.GetByID(strings.TrimSpace(propertyID))
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Properti tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	if property.PartnerID != partner.ID {
		return nil, httpx.Forbidden("Anda tidak berhak mengubah properti ini")
	}
	return property, nil
}

// photoOf mengambil foto dan memastikan kepemilikannya pada properti tertentu.
func (s *MediaService) photoOf(propertyID, photoID string) (*models.PropertyPhoto, error) {
	photo, err := s.deps.Repos.Properties.GetPhoto(strings.TrimSpace(photoID))
	if err != nil {
		if database.IsNotFound(err) {
			return nil, httpx.NotFound("Foto tidak ditemukan")
		}
		return nil, httpx.Internal(err)
	}
	if photo.PropertyID != propertyID {
		return nil, httpx.NotFound("Foto tidak ditemukan pada properti ini")
	}
	return photo, nil
}

// promoteFirst menjadikan foto pertama sebagai foto utama bila belum ada.
func (s *MediaService) promoteFirst(propertyID string) {
	photos, err := s.deps.Repos.Properties.ListPhotos(propertyID)
	if err != nil || len(photos) == 0 {
		return
	}
	for _, photo := range photos {
		if photo.IsPrimary {
			return
		}
	}
	photo := photos[0]
	photo.IsPrimary = true
	_ = s.deps.Repos.Properties.UpdatePhoto(photo)
}

// listViews membentuk daftar tampilan foto properti.
func (s *MediaService) listViews(propertyID string) ([]dto.PhotoView, error) {
	photos, err := s.deps.Repos.Properties.ListPhotos(propertyID)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	views := make([]dto.PhotoView, 0, len(photos))
	for _, photo := range photos {
		views = append(views, dto.NewPhotoView(photo))
	}
	return views, nil
}

// validateFile memeriksa ukuran dan tipe berkas dari isi sesungguhnya.
func (s *MediaService) validateFile(file *multipart.FileHeader) error {
	cfg := s.deps.Config.Storage
	if file == nil {
		return httpx.BadRequest("Berkas foto tidak ditemukan")
	}
	if cfg.MaxUploadBytes > 0 && file.Size > cfg.MaxUploadBytes {
		return httpx.BadRequest("Ukuran foto melebihi batas yang diizinkan")
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
	if !allowedMimeType(mime, cfg.AllowedMimeTypes) {
		return httpx.BadRequest("Tipe berkas " + mime + " tidak diizinkan")
	}
	return nil
}

// allowedMimeType memeriksa apakah tipe berkas termasuk daftar yang diizinkan.
func allowedMimeType(mime string, allowed []string) bool {
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

// readUpload membaca seluruh isi berkas unggahan ke memori.
func readUpload(file *multipart.FileHeader) ([]byte, error) {
	opened, err := file.Open()
	if err != nil {
		return nil, httpx.BadRequest("Berkas foto tidak dapat dibaca")
	}
	defer opened.Close()

	buffer := bytes.NewBuffer(make([]byte, 0, file.Size))
	if _, err := buffer.ReadFrom(opened); err != nil {
		return nil, httpx.BadRequest("Berkas foto gagal dibaca")
	}
	return buffer.Bytes(), nil
}

// processImage mengubah ukuran foto menjadi paling lebar maxPhotoWidth dan
// mengembalikan dimensi akhirnya.
func processImage(raw []byte) ([]byte, int, int, error) {
	img, err := imaging.Decode(bytes.NewReader(raw), imaging.AutoOrientation(true))
	if err != nil {
		return nil, 0, 0, httpx.BadRequest("Berkas foto tidak dapat diproses")
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width > maxPhotoWidth {
		img = imaging.Resize(img, maxPhotoWidth, 0, imaging.Lanczos)
		bounds = img.Bounds()
		width, height = bounds.Dx(), bounds.Dy()
	}

	buffer := new(bytes.Buffer)
	if err := imaging.Encode(buffer, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, 0, 0, httpx.Internal(err)
	}
	return buffer.Bytes(), width, height, nil
}
