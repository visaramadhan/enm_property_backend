package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const agreementColumns = `id, version, title, content, checklist, is_active, effective_at, created_by, created_at`

const acceptanceColumns = `id, agreement_id, agreement_version, user_id, partner_id, ` +
	`accepted_at, ip_address, user_agent`

type postgresAgreementRepository struct{ pool pgxPool }

// scanAgreement membaca satu baris perjanjian dari hasil query.
func scanAgreement(row pgx.Row) (*models.Agreement, error) {
	var agreement models.Agreement
	if err := row.Scan(
		&agreement.ID, &agreement.Version, &agreement.Title, &agreement.Content,
		&agreement.Checklist, &agreement.IsActive, &agreement.EffectiveAt,
		&agreement.CreatedBy, &agreement.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &agreement, nil
}

// scanAcceptance membaca satu baris bukti persetujuan dari hasil query.
func scanAcceptance(row pgx.Row) (*models.AgreementAcceptance, error) {
	var acceptance models.AgreementAcceptance
	if err := row.Scan(
		&acceptance.ID, &acceptance.AgreementID, &acceptance.AgreementVersion,
		&acceptance.UserID, &acceptance.PartnerID, &acceptance.AcceptedAt,
		&acceptance.IPAddress, &acceptance.UserAgent,
	); err != nil {
		return nil, err
	}
	return &acceptance, nil
}

func (r *postgresAgreementRepository) Create(agreement *models.Agreement) error {
	ctx := context.Background()
	const query = `INSERT INTO agreements (` + agreementColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.pool.Exec(ctx, query,
		agreement.ID, agreement.Version, agreement.Title, agreement.Content,
		agreement.Checklist, agreement.IsActive, agreement.EffectiveAt,
		agreement.CreatedBy, agreement.CreatedAt,
	)
	return err
}

func (r *postgresAgreementRepository) Update(agreement *models.Agreement) error {
	ctx := context.Background()
	const query = `UPDATE agreements SET
		version = $2, title = $3, content = $4, checklist = $5, is_active = $6,
		effective_at = $7, created_by = $8, created_at = $9
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		agreement.ID, agreement.Version, agreement.Title, agreement.Content,
		agreement.Checklist, agreement.IsActive, agreement.EffectiveAt,
		agreement.CreatedBy, agreement.CreatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresAgreementRepository) GetByID(id string) (*models.Agreement, error) {
	ctx := context.Background()
	agreement, err := scanAgreement(r.pool.QueryRow(ctx, `SELECT `+agreementColumns+` FROM agreements WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return agreement, nil
}

// GetActive mengembalikan perjanjian aktif dengan EffectiveAt terbaru.
func (r *postgresAgreementRepository) GetActive() (*models.Agreement, error) {
	ctx := context.Background()
	const query = `SELECT ` + agreementColumns + ` FROM agreements
		WHERE is_active = true ORDER BY effective_at DESC, created_at DESC LIMIT 1`
	agreement, err := scanAgreement(r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return agreement, nil
}

func (r *postgresAgreementRepository) List(q ListQuery) ([]*models.Agreement, int, error) {
	ctx := context.Background()
	var w whereBuilder

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "agreements", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + agreementColumns + " FROM agreements" + whereSQL +
		" ORDER BY effective_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	agreements := make([]*models.Agreement, 0)
	for rows.Next() {
		agreement, err := scanAgreement(rows)
		if err != nil {
			return nil, 0, err
		}
		agreements = append(agreements, agreement)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return agreements, total, nil
}

func (r *postgresAgreementRepository) SaveAcceptance(acceptance *models.AgreementAcceptance) error {
	ctx := context.Background()
	const query = `INSERT INTO agreement_acceptances (` + acceptanceColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			agreement_id = EXCLUDED.agreement_id,
			agreement_version = EXCLUDED.agreement_version,
			user_id = EXCLUDED.user_id,
			partner_id = EXCLUDED.partner_id,
			accepted_at = EXCLUDED.accepted_at,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent`
	_, err := r.pool.Exec(ctx, query,
		acceptance.ID, acceptance.AgreementID, acceptance.AgreementVersion,
		acceptance.UserID, acceptance.PartnerID, acceptance.AcceptedAt,
		acceptance.IPAddress, acceptance.UserAgent,
	)
	return err
}

func (r *postgresAgreementRepository) GetAcceptance(partnerID, agreementID string) (*models.AgreementAcceptance, error) {
	ctx := context.Background()
	const query = `SELECT ` + acceptanceColumns + ` FROM agreement_acceptances
		WHERE partner_id = $1 AND agreement_id = $2 LIMIT 1`
	acceptance, err := scanAcceptance(r.pool.QueryRow(ctx, query, partnerID, agreementID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return acceptance, nil
}

func (r *postgresAgreementRepository) ListAcceptances(q ListQuery) ([]*models.AgreementAcceptance, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eq("partner_id", q.Filter("partner_id"))
	w.eq("agreement_id", q.Filter("agreement_id"))

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "agreement_acceptances", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + acceptanceColumns + " FROM agreement_acceptances" + whereSQL +
		" ORDER BY accepted_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	acceptances := make([]*models.AgreementAcceptance, 0)
	for rows.Next() {
		acceptance, err := scanAcceptance(rows)
		if err != nil {
			return nil, 0, err
		}
		acceptances = append(acceptances, acceptance)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return acceptances, total, nil
}
