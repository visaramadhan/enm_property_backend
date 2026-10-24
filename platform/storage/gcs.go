package storage

import (
	"context"
	"fmt"
	"strings"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"enmproperty/backend/config"
)

// gcsStorage menyimpan berkas pada bucket Google Cloud Storage milik Firebase.
type gcsStorage struct {
	client *gcs.Client
	bucket string
}

func newGCSStorage(_ config.Storage, firebaseCfg config.Firebase) (Storage, error) {
	if firebaseCfg.StorageBucket == "" {
		return nil, fmt.Errorf("FIREBASE_STORAGE_BUCKET wajib diisi untuk penyimpanan gcs")
	}

	var opts []option.ClientOption
	switch {
	case firebaseCfg.CredentialsFile != "":
		opts = append(opts, option.WithCredentialsFile(firebaseCfg.CredentialsFile))
	case firebaseCfg.CredentialsJSON != "":
		opts = append(opts, option.WithCredentialsJSON([]byte(firebaseCfg.CredentialsJSON)))
	}

	client, err := gcs.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien Cloud Storage: %w", err)
	}

	return &gcsStorage{client: client, bucket: firebaseCfg.StorageBucket}, nil
}

func (s *gcsStorage) Save(ctx context.Context, objectPath string, data []byte) (string, error) {
	rel := normalizePath(objectPath)
	writer := s.client.Bucket(s.bucket).Object(rel).NewWriter(ctx)
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return "", fmt.Errorf("gagal mengunggah berkas media: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("gagal menyelesaikan unggahan media: %w", err)
	}
	return s.PublicURL(rel), nil
}

func (s *gcsStorage) Delete(ctx context.Context, objectPath string) error {
	rel := normalizePath(objectPath)
	if err := s.client.Bucket(s.bucket).Object(rel).Delete(ctx); err != nil {
		if err == gcs.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("gagal menghapus berkas media: %w", err)
	}
	return nil
}

func (s *gcsStorage) PublicURL(objectPath string) string {
	rel := normalizePath(objectPath)
	return "https://storage.googleapis.com/" + strings.TrimRight(s.bucket, "/") + "/" + rel
}
