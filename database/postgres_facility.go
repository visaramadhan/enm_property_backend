package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const facilityColumns = `id, slug, name, description, icon, sort_order, is_active, created_at, updated_at`

type postgresFacilityRepository struct{ pool pgxPool }

// scanFacility membaca satu baris fasilitas dari hasil query.
func scanFacility(row pgx.Row) (*models.Facility, error) {
	var facility models.Facility
	if err := row.Scan(
		&facility.ID, &facility.Slug, &facility.Name, &facility.Description,
		&facility.Icon, &facility.SortOrder, &facility.IsActive,
		&facility.CreatedAt, &facility.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &facility, nil
}

func (r *postgresFacilityRepository) Create(facility *models.Facility) error {
	ctx := context.Background()

	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM facilities WHERE lower(slug) = lower($1)`, facility.Slug).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("slug fasilitas sudah dipakai")
	}

	const query = `INSERT INTO facilities (` + facilityColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.pool.Exec(ctx, query,
		facility.ID, facility.Slug, facility.Name, facility.Description,
		facility.Icon, facility.SortOrder, facility.IsActive,
		facility.CreatedAt, facility.UpdatedAt,
	)
	return err
}

func (r *postgresFacilityRepository) Update(facility *models.Facility) error {
	ctx := context.Background()
	const query = `UPDATE facilities SET
		slug = $2, name = $3, description = $4, icon = $5, sort_order = $6, is_active = $7,
		created_at = $8, updated_at = $9
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		facility.ID, facility.Slug, facility.Name, facility.Description,
		facility.Icon, facility.SortOrder, facility.IsActive,
		facility.CreatedAt, facility.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresFacilityRepository) Delete(id string) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `DELETE FROM facilities WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresFacilityRepository) GetByID(id string) (*models.Facility, error) {
	ctx := context.Background()
	facility, err := scanFacility(r.pool.QueryRow(ctx, `SELECT `+facilityColumns+` FROM facilities WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return facility, nil
}

func (r *postgresFacilityRepository) GetBySlug(slug string) (*models.Facility, error) {
	ctx := context.Background()
	facility, err := scanFacility(r.pool.QueryRow(ctx, `SELECT `+facilityColumns+` FROM facilities WHERE lower(slug) = lower($1) LIMIT 1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return facility, nil
}

func (r *postgresFacilityRepository) SlugExists(slug, excludeID string) bool {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM facilities WHERE lower(slug) = lower($1) AND id <> $2`
	if err := r.pool.QueryRow(ctx, query, slug, excludeID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (r *postgresFacilityRepository) List(q ListQuery) ([]*models.Facility, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.boolFilter("is_active", q.Filter("is_active"))
	w.searchOr([]string{"name", "description"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "facilities", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + facilityColumns + " FROM facilities" + whereSQL +
		" ORDER BY sort_order ASC, lower(name) ASC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	facilities := make([]*models.Facility, 0)
	for rows.Next() {
		facility, err := scanFacility(rows)
		if err != nil {
			return nil, 0, err
		}
		facilities = append(facilities, facility)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return facilities, total, nil
}
