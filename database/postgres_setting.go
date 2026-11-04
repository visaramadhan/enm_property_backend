package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const settingColumns = `key, value, description, is_secret, updated_by, updated_at`

type postgresSettingRepository struct{ pool pgxPool }

// scanSetting membaca satu baris pengaturan dari hasil query.
func scanSetting(row pgx.Row) (*models.Setting, error) {
	var setting models.Setting
	if err := row.Scan(
		&setting.Key, &setting.Value, &setting.Description, &setting.IsSecret,
		&setting.UpdatedBy, &setting.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *postgresSettingRepository) All() []*models.Setting {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT `+settingColumns+` FROM settings ORDER BY key ASC`)
	if err != nil {
		return []*models.Setting{}
	}
	defer rows.Close()

	settings := make([]*models.Setting, 0)
	for rows.Next() {
		setting, err := scanSetting(rows)
		if err != nil {
			return []*models.Setting{}
		}
		settings = append(settings, setting)
	}
	return settings
}

func (r *postgresSettingRepository) Get(key string) (*models.Setting, error) {
	ctx := context.Background()
	setting, err := scanSetting(r.pool.QueryRow(ctx, `SELECT `+settingColumns+` FROM settings WHERE key = $1`, key))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return setting, nil
}

// Set menyimpan pengaturan dengan mempertahankan deskripsi dan status rahasia
// lama bila nilai baru tidak mengisinya (sama seperti implementasi memori).
func (r *postgresSettingRepository) Set(setting *models.Setting) error {
	ctx := context.Background()
	const query = `INSERT INTO settings (` + settingColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			description = CASE WHEN EXCLUDED.description = '' THEN settings.description
				ELSE EXCLUDED.description END,
			is_secret = settings.is_secret OR EXCLUDED.is_secret,
			updated_by = EXCLUDED.updated_by,
			updated_at = EXCLUDED.updated_at`
	_, err := r.pool.Exec(ctx, query,
		setting.Key, setting.Value, setting.Description, setting.IsSecret,
		setting.UpdatedBy, setting.UpdatedAt,
	)
	return err
}
