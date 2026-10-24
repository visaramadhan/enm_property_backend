package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"enmproperty/backend/config"
)

// localStorage menyimpan berkas pada disk dan menyajikannya lewat prefix publik.
type localStorage struct {
	dir     string
	baseURL string
}

func newLocalStorage(cfg config.Storage) (Storage, error) {
	if err := os.MkdirAll(cfg.LocalDir, 0o755); err != nil {
		return nil, fmt.Errorf("gagal menyiapkan direktori storage: %w", err)
	}
	return &localStorage{
		dir:     cfg.LocalDir,
		baseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (s *localStorage) Save(_ context.Context, objectPath string, data []byte) (string, error) {
	rel := normalizePath(objectPath)
	full := filepath.Join(s.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", fmt.Errorf("gagal menyiapkan folder media: %w", err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", fmt.Errorf("gagal menulis berkas media: %w", err)
	}
	return s.PublicURL(rel), nil
}

func (s *localStorage) Delete(_ context.Context, objectPath string) error {
	rel := normalizePath(objectPath)
	full := filepath.Join(s.dir, filepath.FromSlash(rel))
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("gagal menghapus berkas media: %w", err)
	}
	return nil
}

func (s *localStorage) PublicURL(objectPath string) string {
	rel := normalizePath(objectPath)
	if s.baseURL == "" {
		return "/" + rel
	}
	return s.baseURL + "/" + rel
}
