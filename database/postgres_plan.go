package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const planColumns = `id, slug, name, description, price, duration_days, listing_quota, features, ` +
	`is_free, is_active, sort_order, level, boost_enabled, boost_quota, keyword_access, ` +
	`catalog_access, whatsapp_access, created_at, updated_at`

type postgresPlanRepository struct{ pool pgxPool }

// scanPlan membaca satu baris paket dari hasil query.
func scanPlan(row pgx.Row) (*models.Plan, error) {
	var plan models.Plan
	if err := row.Scan(
		&plan.ID, &plan.Slug, &plan.Name, &plan.Description, &plan.Price, &plan.DurationDays,
		&plan.ListingQuota, &plan.Features, &plan.IsFree, &plan.IsActive, &plan.SortOrder,
		&plan.Level, &plan.BoostEnabled, &plan.BoostQuota, &plan.KeywordAccess,
		&plan.CatalogAccess, &plan.WhatsAppAccess, &plan.CreatedAt, &plan.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *postgresPlanRepository) Create(plan *models.Plan) error {
	ctx := context.Background()

	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM plans WHERE slug = $1`, plan.Slug).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("slug paket sudah dipakai")
	}

	const query = `INSERT INTO plans (` + planColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`
	_, err := r.pool.Exec(ctx, query,
		plan.ID, plan.Slug, plan.Name, plan.Description, plan.Price, plan.DurationDays,
		plan.ListingQuota, plan.Features, plan.IsFree, plan.IsActive, plan.SortOrder,
		plan.Level, plan.BoostEnabled, plan.BoostQuota, plan.KeywordAccess,
		plan.CatalogAccess, plan.WhatsAppAccess, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *postgresPlanRepository) Update(plan *models.Plan) error {
	ctx := context.Background()
	const query = `UPDATE plans SET
		slug = $2, name = $3, description = $4, price = $5, duration_days = $6,
		listing_quota = $7, features = $8, is_free = $9, is_active = $10, sort_order = $11,
		level = $12, boost_enabled = $13, boost_quota = $14, keyword_access = $15,
		catalog_access = $16, whatsapp_access = $17, created_at = $18, updated_at = $19
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		plan.ID, plan.Slug, plan.Name, plan.Description, plan.Price, plan.DurationDays,
		plan.ListingQuota, plan.Features, plan.IsFree, plan.IsActive, plan.SortOrder,
		plan.Level, plan.BoostEnabled, plan.BoostQuota, plan.KeywordAccess,
		plan.CatalogAccess, plan.WhatsAppAccess, plan.CreatedAt, plan.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresPlanRepository) GetByID(id string) (*models.Plan, error) {
	ctx := context.Background()
	plan, err := scanPlan(r.pool.QueryRow(ctx, `SELECT `+planColumns+` FROM plans WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return plan, nil
}

func (r *postgresPlanRepository) GetBySlug(slug string) (*models.Plan, error) {
	ctx := context.Background()
	plan, err := scanPlan(r.pool.QueryRow(ctx, `SELECT `+planColumns+` FROM plans WHERE slug = $1 LIMIT 1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return plan, nil
}

func (r *postgresPlanRepository) List(q ListQuery) ([]*models.Plan, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.boolFilter("is_active", q.Filter("is_active"))
	w.contains("name", q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "plans", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + planColumns + " FROM plans" + whereSQL + " ORDER BY sort_order ASC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	plans := make([]*models.Plan, 0)
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}
