package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"enmproperty/backend/config"
)

// pgxPool adalah irisan method *pgxpool.Pool yang dipakai seluruh repository.
// Ketergantungan pada antarmuka ini memudahkan pengujian dan menjaga tiap
// berkas repository tetap ringkas.
type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// NewPostgresPool menyiapkan pool koneksi PostgreSQL (pgxpool) dari konfigurasi
// dan memastikan koneksi dapat dijangkau melalui Ping.
func NewPostgresPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("DB_URL tidak valid: %w", err)
	}
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool PostgreSQL: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke PostgreSQL: %w", err)
	}
	return pool, nil
}

// whereBuilder menyusun klausa WHERE beserta argumen berparameter posisi.
//
// Seluruh nilai dikirim sebagai parameter ($1, $2, ...) sehingga aman dari
// penyisipan SQL, sedangkan nama kolom selalu berasal dari konstanta internal.
type whereBuilder struct {
	conds []string
	args  []any
}

// arg menambahkan satu argumen dan mengembalikan placeholder posisinya.
func (w *whereBuilder) arg(value any) string {
	w.args = append(w.args, value)
	return fmt.Sprintf("$%d", len(w.args))
}

// eq menambahkan filter kesetaraan bila nilainya tidak kosong.
func (w *whereBuilder) eq(column, value string) {
	if value == "" {
		return
	}
	w.conds = append(w.conds, column+" = "+w.arg(value))
}

// eqFold menambahkan filter kesetaraan tanpa memandang huruf besar-kecil.
func (w *whereBuilder) eqFold(column, value string) {
	if value == "" {
		return
	}
	w.conds = append(w.conds, "lower("+column+") = lower("+w.arg(value)+")")
}

// contains menambahkan filter "mengandung" (setara fungsi ContainsFold).
func (w *whereBuilder) contains(column, value string) {
	if value == "" {
		return
	}
	w.conds = append(w.conds, "strpos(lower("+column+"), lower("+w.arg(value)+")) > 0")
}

// searchOr menambahkan pencarian "mengandung" pada beberapa kolom sekaligus.
func (w *whereBuilder) searchOr(columns []string, value string) {
	if value == "" {
		return
	}
	placeholder := w.arg(value)
	parts := make([]string, len(columns))
	for i, column := range columns {
		parts[i] = "strpos(lower(" + column + "), lower(" + placeholder + ")) > 0"
	}
	w.conds = append(w.conds, "("+strings.Join(parts, " OR ")+")")
}

// boolFilter menerapkan filter boolean opsional dari ListQuery ("true"/"false").
func (w *whereBuilder) boolFilter(column, value string) {
	switch value {
	case "true":
		w.conds = append(w.conds, column+" = true")
	case "false":
		w.conds = append(w.conds, column+" = false")
	}
}

// isNull menambahkan filter kolom bernilai NULL.
func (w *whereBuilder) isNull(column string) {
	w.conds = append(w.conds, column+" IS NULL")
}

// isNotNull menambahkan filter kolom tidak bernilai NULL.
func (w *whereBuilder) isNotNull(column string) {
	w.conds = append(w.conds, column+" IS NOT NULL")
}

// timeRange menambahkan filter rentang waktu dari kunci "from" dan "to" (RFC3339).
func (w *whereBuilder) timeRange(column string, q ListQuery) {
	if from := q.Filter("from"); from != "" {
		if parsed, err := time.Parse(time.RFC3339, from); err == nil {
			w.conds = append(w.conds, column+" >= "+w.arg(parsed))
		}
	}
	if to := q.Filter("to"); to != "" {
		if parsed, err := time.Parse(time.RFC3339, to); err == nil {
			w.conds = append(w.conds, column+" <= "+w.arg(parsed))
		}
	}
}

// where mengembalikan klausa WHERE lengkap, atau string kosong bila tanpa filter.
func (w *whereBuilder) where() string {
	if len(w.conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.conds, " AND ")
}

// paginate menambahkan LIMIT/OFFSET bila paginasi aktif.
func (w *whereBuilder) paginate(q ListQuery) string {
	if q.PerPage <= 0 {
		return ""
	}
	return " LIMIT " + w.arg(q.PerPage) + " OFFSET " + w.arg(q.Offset())
}

// countRows menghitung jumlah baris pada tabel dengan klausa WHERE yang sama.
func countRows(ctx context.Context, pool pgxPool, table, whereSQL string, args []any) (int, error) {
	var total int
	query := "SELECT count(*) FROM " + table + whereSQL
	if err := pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// orderClause memilih klausa ORDER BY berdasarkan whitelist kolom agar aman.
func orderClause(field, defaultColumn string, desc bool, allowed map[string]string) string {
	column, ok := allowed[field]
	if !ok {
		column = defaultColumn
	}
	direction := " ASC"
	if desc {
		direction = " DESC"
	}
	return column + direction
}
