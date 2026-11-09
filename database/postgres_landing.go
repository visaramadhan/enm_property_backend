package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const landingColumns = `key, value, updated_by, updated_at`

type postgresLandingRepository struct{ pool pgxPool }

// scanLanding membaca satu baris konten landing dari hasil query.
func scanLanding(row pgx.Row) (*models.LandingContent, error) {
	var content models.LandingContent
	if err := row.Scan(&content.Key, &content.Value, &content.UpdatedBy, &content.UpdatedAt); err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *postgresLandingRepository) All() []*models.LandingContent {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT `+landingColumns+` FROM landing_content ORDER BY key ASC`)
	if err != nil {
		return []*models.LandingContent{}
	}
	defer rows.Close()

	items := make([]*models.LandingContent, 0)
	for rows.Next() {
		item, err := scanLanding(rows)
		if err != nil {
			return []*models.LandingContent{}
		}
		items = append(items, item)
	}
	return items
}

func (r *postgresLandingRepository) Get(key string) (*models.LandingContent, error) {
	ctx := context.Background()
	item, err := scanLanding(r.pool.QueryRow(ctx, `SELECT `+landingColumns+` FROM landing_content WHERE key = $1`, key))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

// Set menyimpan konten landing dengan pola upsert kunci-nilai.
func (r *postgresLandingRepository) Set(content *models.LandingContent) error {
	ctx := context.Background()
	const query = `INSERT INTO landing_content (` + landingColumns + `)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_by = EXCLUDED.updated_by,
			updated_at = EXCLUDED.updated_at`
	_, err := r.pool.Exec(ctx, query,
		content.Key, content.Value, content.UpdatedBy, content.UpdatedAt,
	)
	return err
}
