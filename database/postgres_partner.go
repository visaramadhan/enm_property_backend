package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const partnerColumns = `id, user_id, business_name, pic_name, phone, email, address, city, ` +
	`province, status, suspended_reason, agreement_accepted_at, joined_at, created_at, updated_at`

type postgresPartnerRepository struct{ pool pgxPool }

// scanPartner membaca satu baris mitra dari hasil query.
func scanPartner(row pgx.Row) (*models.Partner, error) {
	var partner models.Partner
	if err := row.Scan(
		&partner.ID, &partner.UserID, &partner.BusinessName, &partner.PICName, &partner.Phone,
		&partner.Email, &partner.Address, &partner.City, &partner.Province, &partner.Status,
		&partner.SuspendedReason, &partner.AgreementAcceptedAt, &partner.JoinedAt,
		&partner.CreatedAt, &partner.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *postgresPartnerRepository) Create(partner *models.Partner) error {
	ctx := context.Background()
	const query = `INSERT INTO partners (` + partnerColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`
	_, err := r.pool.Exec(ctx, query,
		partner.ID, partner.UserID, partner.BusinessName, partner.PICName, partner.Phone,
		partner.Email, partner.Address, partner.City, partner.Province, string(partner.Status),
		partner.SuspendedReason, partner.AgreementAcceptedAt, partner.JoinedAt,
		partner.CreatedAt, partner.UpdatedAt,
	)
	return err
}

func (r *postgresPartnerRepository) Update(partner *models.Partner) error {
	ctx := context.Background()
	const query = `UPDATE partners SET
		user_id = $2, business_name = $3, pic_name = $4, phone = $5, email = $6,
		address = $7, city = $8, province = $9, status = $10, suspended_reason = $11,
		agreement_accepted_at = $12, joined_at = $13, created_at = $14, updated_at = $15
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		partner.ID, partner.UserID, partner.BusinessName, partner.PICName, partner.Phone,
		partner.Email, partner.Address, partner.City, partner.Province, string(partner.Status),
		partner.SuspendedReason, partner.AgreementAcceptedAt, partner.JoinedAt,
		partner.CreatedAt, partner.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresPartnerRepository) GetByID(id string) (*models.Partner, error) {
	ctx := context.Background()
	partner, err := scanPartner(r.pool.QueryRow(ctx, `SELECT `+partnerColumns+` FROM partners WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return partner, nil
}

func (r *postgresPartnerRepository) GetByUserID(userID string) (*models.Partner, error) {
	ctx := context.Background()
	partner, err := scanPartner(r.pool.QueryRow(ctx, `SELECT `+partnerColumns+` FROM partners WHERE user_id = $1 LIMIT 1`, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return partner, nil
}

func (r *postgresPartnerRepository) List(q ListQuery) ([]*models.Partner, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eq("status", q.Filter("status"))
	w.eqFold("city", q.Filter("city"))
	w.searchOr([]string{"business_name", "pic_name", "email"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "partners", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + partnerColumns + " FROM partners" + whereSQL + " ORDER BY created_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	partners := make([]*models.Partner, 0)
	for rows.Next() {
		partner, err := scanPartner(rows)
		if err != nil {
			return nil, 0, err
		}
		partners = append(partners, partner)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return partners, total, nil
}
