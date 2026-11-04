package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const subscriptionColumns = `id, partner_id, plan_id, plan_name, order_id, starts_at, ends_at, ` +
	`status, created_at, updated_at`

type postgresSubscriptionRepository struct{ pool pgxPool }

// scanSubscription membaca satu baris langganan dari hasil query.
func scanSubscription(row pgx.Row) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := row.Scan(
		&subscription.ID, &subscription.PartnerID, &subscription.PlanID, &subscription.PlanName,
		&subscription.OrderID, &subscription.StartsAt, &subscription.EndsAt,
		&subscription.Status, &subscription.CreatedAt, &subscription.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *postgresSubscriptionRepository) Create(subscription *models.Subscription) error {
	ctx := context.Background()
	const query = `INSERT INTO subscriptions (` + subscriptionColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, query,
		subscription.ID, subscription.PartnerID, subscription.PlanID, subscription.PlanName,
		subscription.OrderID, subscription.StartsAt, subscription.EndsAt,
		string(subscription.Status), subscription.CreatedAt, subscription.UpdatedAt,
	)
	return err
}

func (r *postgresSubscriptionRepository) Update(subscription *models.Subscription) error {
	ctx := context.Background()
	const query = `UPDATE subscriptions SET
		partner_id = $2, plan_id = $3, plan_name = $4, order_id = $5, starts_at = $6,
		ends_at = $7, status = $8, created_at = $9, updated_at = $10
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		subscription.ID, subscription.PartnerID, subscription.PlanID, subscription.PlanName,
		subscription.OrderID, subscription.StartsAt, subscription.EndsAt,
		string(subscription.Status), subscription.CreatedAt, subscription.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresSubscriptionRepository) GetByID(id string) (*models.Subscription, error) {
	ctx := context.Background()
	subscription, err := scanSubscription(r.pool.QueryRow(ctx, `SELECT `+subscriptionColumns+` FROM subscriptions WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return subscription, nil
}

// GetActiveByPartner mengembalikan langganan aktif (status active dan EndsAt
// belum lewat) terbaru milik mitra.
func (r *postgresSubscriptionRepository) GetActiveByPartner(partnerID string) (*models.Subscription, error) {
	ctx := context.Background()
	const query = `SELECT ` + subscriptionColumns + ` FROM subscriptions
		WHERE partner_id = $1 AND status = $2 AND ends_at >= now()
		ORDER BY ends_at DESC, created_at DESC LIMIT 1`
	subscription, err := scanSubscription(r.pool.QueryRow(ctx, query, partnerID, string(models.SubscriptionStatusActive)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return subscription, nil
}

// GetLatestByPartner mengembalikan langganan terbaru mitra tanpa memandang status.
func (r *postgresSubscriptionRepository) GetLatestByPartner(partnerID string) (*models.Subscription, error) {
	ctx := context.Background()
	const query = `SELECT ` + subscriptionColumns + ` FROM subscriptions
		WHERE partner_id = $1 ORDER BY created_at DESC LIMIT 1`
	subscription, err := scanSubscription(r.pool.QueryRow(ctx, query, partnerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return subscription, nil
}

func (r *postgresSubscriptionRepository) List(q ListQuery) ([]*models.Subscription, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eq("partner_id", q.Filter("partner_id"))
	w.eq("plan_id", q.Filter("plan_id"))
	w.eq("status", q.Filter("status"))
	w.timeRange("created_at", q)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "subscriptions", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + subscriptionColumns + " FROM subscriptions" + whereSQL +
		" ORDER BY created_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subscriptions := make([]*models.Subscription, 0)
	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, 0, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return subscriptions, total, nil
}
