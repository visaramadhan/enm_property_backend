// Package storage menyediakan penyimpanan foto properti dengan dua implementasi:
// lokal untuk pengembangan dan Google Cloud Storage untuk produksi.
package storage

import (
	"context"
	"fmt"
	"path"
	"strings"

	"enmproperty/backend/config"
)

// Storage adalah kontrak penyimpanan berkas media.
type Storage interface {
	// Save menyimpan data dan mengembalikan path relatif serta URL publiknya.
	Save(ctx context.Context, objectPath string, data []byte) (string, error)
	// Delete menghapus berkas pada path relatif.
	Delete(ctx context.Context, objectPath string) error
	// PublicURL membentuk URL publik dari path relatif.
	PublicURL(objectPath string) string
}

// New membuat implementasi Storage sesuai STORAGE_DRIVER. Konfigurasi Firebase
// dipakai oleh driver gcs untuk mengetahui bucket dan kredensial.
func New(cfg config.Storage, firebaseCfg config.Firebase) (Storage, error) {
	switch strings.ToLower(cfg.Driver) {
	case "gcs":
		return newGCSStorage(cfg, firebaseCfg)
	case "local", "":
		return newLocalStorage(cfg)
	default:
		return nil, fmt.Errorf("driver storage %q tidak dikenal", cfg.Driver)
	}
}

// normalizePath merapikan path objek agar tidak keluar dari direktori akar.
func normalizePath(objectPath string) string {
	cleaned := path.Clean("/" + strings.ReplaceAll(objectPath, "\\", "/"))
	return strings.TrimPrefix(cleaned, "/")
}
