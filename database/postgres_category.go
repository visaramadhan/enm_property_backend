package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const categoryColumns = `id, slug, name, description, sort_order, is_active, created_at, updated_at`

type postgresCategoryRepository struct{ pool pgxPool }

// scanCategory membaca satu baris kategori dari hasil query.
func scanCategory(row pgx.Row) (*models.Category, error) {
	var category models.Category
	if err := row.Scan(
		&category.ID, &category.Slug, &category.Name, &category.Description,
		&category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *postgresCategoryRepository) Create(category *models.Category) error {
	ctx := context.Background()

	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM categories WHERE lower(slug) = lower($1)`, category.Slug).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("slug kategori sudah dipakai")
	}

	const query = `INSERT INTO categories (` + categoryColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, query,
		category.ID, category.Slug, category.Name, category.Description,
		category.SortOrder, category.IsActive, category.CreatedAt, category.UpdatedAt,
	)
	return err
}

func (r *postgresCategoryRepository) Update(category *models.Category) error {
	ctx := context.Background()
	const query = `UPDATE categories SET
		slug = $2, name = $3, description = $4, sort_order = $5, is_active = $6,
		created_at = $7, updated_at = $8
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		category.ID, category.Slug, category.Name, category.Description,
		category.SortOrder, category.IsActive, category.CreatedAt, category.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresCategoryRepository) Delete(id string) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresCategoryRepository) GetByID(id string) (*models.Category, error) {
	ctx := context.Background()
	category, err := scanCategory(r.pool.QueryRow(ctx, `SELECT `+categoryColumns+` FROM categories WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

func (r *postgresCategoryRepository) GetBySlug(slug string) (*models.Category, error) {
	ctx := context.Background()
	category, err := scanCategory(r.pool.QueryRow(ctx, `SELECT `+categoryColumns+` FROM categories WHERE lower(slug) = lower($1) LIMIT 1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

func (r *postgresCategoryRepository) SlugExists(slug, excludeID string) bool {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM categories WHERE lower(slug) = lower($1) AND id <> $2`
	if err := r.pool.QueryRow(ctx, query, slug, excludeID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (r *postgresCategoryRepository) List(q ListQuery) ([]*models.Category, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.boolFilter("is_active", q.Filter("is_active"))
	w.searchOr([]string{"name", "description"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "categories", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + categoryColumns + " FROM categories" + whereSQL +
		" ORDER BY sort_order ASC, lower(name) ASC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	categories := make([]*models.Category, 0)
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, 0, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return categories, total, nil
}
