package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"enmproperty/backend/models"
)

const orderColumns = `id, code, user_id, partner_id, partner_name, plan_id, plan_name, ` +
	`plan_price, plan_duration_days, plan_listing_quota, amount, kind, status, provider, ` +
	`snap_token, snap_redirect_url, midtrans_transaction_id, midtrans_payment_type, ` +
	`payment_method, payment_channel, payment_number, payment_qr_url, payment_qr_string, ` +
	`payment_expires_at, paid_at, expires_at, created_at, updated_at`

type postgresOrderRepository struct{ pool pgxPool }

// scanOrder membaca satu baris order dari hasil query.
func scanOrder(row pgx.Row) (*models.Order, error) {
	var order models.Order
	if err := row.Scan(
		&order.ID, &order.Code, &order.UserID, &order.PartnerID, &order.PartnerName,
		&order.PlanID, &order.PlanName, &order.PlanPrice, &order.PlanDurationDays,
		&order.PlanListingQuota, &order.Amount, &order.Kind, &order.Status, &order.Provider,
		&order.SnapToken, &order.SnapRedirectURL, &order.MidtransTransactionID,
		&order.MidtransPaymentType, &order.PaymentMethod, &order.PaymentChannel,
		&order.PaymentNumber, &order.PaymentQRURL, &order.PaymentQRString, &order.PaymentExpiresAt,
		&order.PaidAt, &order.ExpiresAt, &order.CreatedAt, &order.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &order, nil
}

// isUniqueViolation memeriksa galat pelanggaran constraint unik PostgreSQL.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *postgresOrderRepository) Create(order *models.Order) error {
	ctx := context.Background()

	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE code = $1`, order.Code).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errDuplicateCode
	}

	const query = `INSERT INTO orders (` + orderColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`
	_, err := r.pool.Exec(ctx, query,
		order.ID, order.Code, order.UserID, order.PartnerID, order.PartnerName,
		order.PlanID, order.PlanName, order.PlanPrice, order.PlanDurationDays,
		order.PlanListingQuota, order.Amount, string(order.Kind), string(order.Status),
		string(order.Provider), order.SnapToken, order.SnapRedirectURL,
		order.MidtransTransactionID, order.MidtransPaymentType,
		order.PaymentMethod, order.PaymentChannel, order.PaymentNumber,
		order.PaymentQRURL, order.PaymentQRString, order.PaymentExpiresAt,
		order.PaidAt, order.ExpiresAt, order.CreatedAt, order.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return errDuplicateCode
	}
	return err
}

func (r *postgresOrderRepository) Update(order *models.Order) error {
	ctx := context.Background()
	const query = `UPDATE orders SET
		code = $2, user_id = $3, partner_id = $4, partner_name = $5, plan_id = $6,
		plan_name = $7, plan_price = $8, plan_duration_days = $9, plan_listing_quota = $10,
		amount = $11, kind = $12, status = $13, provider = $14, snap_token = $15,
		snap_redirect_url = $16, midtrans_transaction_id = $17, midtrans_payment_type = $18,
		payment_method = $19, payment_channel = $20, payment_number = $21, payment_qr_url = $22,
		payment_qr_string = $23, payment_expires_at = $24,
		paid_at = $25, expires_at = $26, created_at = $27, updated_at = $28
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		order.ID, order.Code, order.UserID, order.PartnerID, order.PartnerName,
		order.PlanID, order.PlanName, order.PlanPrice, order.PlanDurationDays,
		order.PlanListingQuota, order.Amount, string(order.Kind), string(order.Status),
		string(order.Provider), order.SnapToken, order.SnapRedirectURL,
		order.MidtransTransactionID, order.MidtransPaymentType,
		order.PaymentMethod, order.PaymentChannel, order.PaymentNumber,
		order.PaymentQRURL, order.PaymentQRString, order.PaymentExpiresAt,
		order.PaidAt, order.ExpiresAt, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresOrderRepository) GetByID(id string) (*models.Order, error) {
	ctx := context.Background()
	order, err := scanOrder(r.pool.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return order, nil
}

func (r *postgresOrderRepository) GetByCode(code string) (*models.Order, error) {
	ctx := context.Background()
	order, err := scanOrder(r.pool.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders WHERE lower(code) = lower($1) LIMIT 1`, code))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return order, nil
}

// CountByCodePrefix menghitung order yang kodenya diawali prefix (peka huruf).
func (r *postgresOrderRepository) CountByCodePrefix(prefix string) int {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM orders WHERE left(code, length($1)) = $1`
	if err := r.pool.QueryRow(ctx, query, prefix).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (r *postgresOrderRepository) List(q ListQuery) ([]*models.Order, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eq("status", q.Filter("status"))
	w.eq("partner_id", q.Filter("partner_id"))
	w.eq("user_id", q.Filter("user_id"))
	w.eq("plan_id", q.Filter("plan_id"))
	w.eq("provider", q.Filter("provider"))
	w.timeRange("created_at", q)
	w.searchOr([]string{"code", "partner_name"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "orders", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + orderColumns + " FROM orders" + whereSQL +
		" ORDER BY created_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*models.Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
