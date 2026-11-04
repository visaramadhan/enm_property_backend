package database

import (
	"context"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const auditColumns = `id, actor_id, actor_name, actor_email, actor_role, action, entity, ` +
	`entity_id, "before", "after", note, ip_address, user_agent, created_at`

type postgresAuditRepository struct{ pool pgxPool }

// scanAudit membaca satu baris catatan audit dari hasil query.
func scanAudit(row pgx.Row) (*models.AuditLog, error) {
	var entry models.AuditLog
	if err := row.Scan(
		&entry.ID, &entry.ActorID, &entry.ActorName, &entry.ActorEmail, &entry.ActorRole,
		&entry.Action, &entry.Entity, &entry.EntityID, &entry.Before, &entry.After,
		&entry.Note, &entry.IPAddress, &entry.UserAgent, &entry.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *postgresAuditRepository) Create(entry *models.AuditLog) error {
	ctx := context.Background()
	const query = `INSERT INTO audit_logs (` + auditColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.ActorID, entry.ActorName, entry.ActorEmail, string(entry.ActorRole),
		string(entry.Action), entry.Entity, entry.EntityID, entry.Before, entry.After,
		entry.Note, entry.IPAddress, entry.UserAgent, entry.CreatedAt,
	)
	return err
}

func (r *postgresAuditRepository) List(q ListQuery) ([]*models.AuditLog, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eqFold("entity", q.Filter("entity"))
	w.eq("entity_id", q.Filter("entity_id"))
	w.eq("actor_id", q.Filter("actor_id"))
	w.eq("actor_role", q.Filter("actor_role"))
	w.eq("action", q.Filter("action"))
	w.timeRange("created_at", q)
	w.searchOr([]string{"actor_name", "actor_email", "note"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "audit_logs", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT " + auditColumns + " FROM audit_logs" + whereSQL +
		" ORDER BY created_at DESC" + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := make([]*models.AuditLog, 0)
	for rows.Next() {
		entry, err := scanAudit(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func (r *postgresAuditRepository) All() []*models.AuditLog {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT `+auditColumns+` FROM audit_logs ORDER BY created_at ASC`)
	if err != nil {
		return []*models.AuditLog{}
	}
	defer rows.Close()

	entries := make([]*models.AuditLog, 0)
	for rows.Next() {
		entry, err := scanAudit(rows)
		if err != nil {
			return []*models.AuditLog{}
		}
		entries = append(entries, entry)
	}
	return entries
}
