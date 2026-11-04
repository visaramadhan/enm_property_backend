// Package migrations menyediakan runner migrasi SQL ringan (up-only).
//
// Berkas .sql di-embed ke dalam biner sehingga tidak bergantung pada berkas
// eksternal maupun DB_MIGRATIONS_PATH. Setiap berkas dieksekusi berurutan di
// dalam satu transaksi dan versinya dicatat pada tabel schema_migrations,
// sehingga pemanggilan berulang bersifat idempoten.
package migrations

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

// Run menerapkan seluruh migrasi yang belum pernah dijalankan.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ensureSchemaTable(ctx, pool); err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		content, err := files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("gagal membaca migrasi %s: %w", name, err)
		}
		if err := apply(ctx, pool, version, string(content)); err != nil {
			return fmt.Errorf("gagal menerapkan migrasi %s: %w", name, err)
		}
	}
	return nil
}

// migrationNames mengembalikan nama berkas .sql dalam urutan menaik.
func migrationNames() ([]string, error) {
	entries, err := files.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("gagal membaca direktori migrasi: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

// ensureSchemaTable membuat tabel pencatat versi migrasi bila belum ada.
func ensureSchemaTable(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text        PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("gagal menyiapkan tabel schema_migrations: %w", err)
	}
	return nil
}

// appliedVersions memuat daftar versi yang sudah diterapkan.
func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca versi migrasi: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("gagal membaca versi migrasi: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca versi migrasi: %w", err)
	}
	return applied, nil
}

// apply menjalankan satu berkas migrasi beserta pencatatan versinya.
func apply(ctx context.Context, pool *pgxpool.Pool, version, content string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, content); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
