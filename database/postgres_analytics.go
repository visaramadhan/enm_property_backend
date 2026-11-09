package database

import (
	"context"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const viewColumns = `id, property_id, partner_id, viewed_at`
const contactColumns = `id, property_id, partner_id, user_id, channel, contact_name, ` +
	`contact_phone, message, created_at`

type postgresAnalyticsRepository struct{ pool pgxPool }

// scanView membaca satu baris tayangan dari hasil query.
func scanView(row pgx.Row) (*models.PropertyView, error) {
	var view models.PropertyView
	if err := row.Scan(&view.ID, &view.PropertyID, &view.PartnerID, &view.ViewedAt); err != nil {
		return nil, err
	}
	return &view, nil
}

// scanContact membaca satu baris lead dari hasil query.
func scanContact(row pgx.Row) (*models.PropertyContact, error) {
	var contact models.PropertyContact
	if err := row.Scan(
		&contact.ID, &contact.PropertyID, &contact.PartnerID, &contact.UserID,
		&contact.Channel, &contact.Name, &contact.Phone, &contact.Message,
		&contact.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &contact, nil
}

// AddView mencatat tayangan sekaligus menaikkan penghitung pada properti.
func (r *postgresAnalyticsRepository) AddView(view *models.PropertyView) error {
	ctx := context.Background()
	const query = `INSERT INTO property_views (` + viewColumns + `) VALUES ($1,$2,$3,$4)`
	if _, err := r.pool.Exec(ctx, query, view.ID, view.PropertyID, view.PartnerID, view.ViewedAt); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE properties SET views = views + 1 WHERE id = $1`, view.PropertyID)
	return err
}

// AddContact mencatat lead sekaligus menaikkan penghitung pada properti.
func (r *postgresAnalyticsRepository) AddContact(contact *models.PropertyContact) error {
	ctx := context.Background()
	const query = `INSERT INTO property_contacts (` + contactColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	if _, err := r.pool.Exec(ctx, query,
		contact.ID, contact.PropertyID, contact.PartnerID, contact.UserID,
		contact.Channel, contact.Name, contact.Phone, contact.Message, contact.CreatedAt,
	); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE properties SET contact_count = contact_count + 1 WHERE id = $1`, contact.PropertyID)
	return err
}

// listContacts menjalankan query lead dengan klausa penyaring apa pun.
func (r *postgresAnalyticsRepository) listContacts(where string, arg string) ([]*models.PropertyContact, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT `+contactColumns+` FROM property_contacts WHERE `+where+` ORDER BY created_at DESC`, arg)
	if err != nil {
		return []*models.PropertyContact{}, err
	}
	defer rows.Close()

	contacts := make([]*models.PropertyContact, 0)
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return []*models.PropertyContact{}, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// ListContactsByProperty mengembalikan lead pada satu properti.
func (r *postgresAnalyticsRepository) ListContactsByProperty(propertyID string) ([]*models.PropertyContact, error) {
	return r.listContacts("property_id = $1", propertyID)
}

// ContactsByPartner mengembalikan seluruh lead milik satu mitra.
func (r *postgresAnalyticsRepository) ContactsByPartner(partnerID string) ([]*models.PropertyContact, error) {
	return r.listContacts("partner_id = $1", partnerID)
}

func (r *postgresAnalyticsRepository) Views() []*models.PropertyView {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT `+viewColumns+` FROM property_views ORDER BY viewed_at ASC`)
	if err != nil {
		return []*models.PropertyView{}
	}
	defer rows.Close()

	views := make([]*models.PropertyView, 0)
	for rows.Next() {
		view, err := scanView(rows)
		if err != nil {
			return []*models.PropertyView{}
		}
		views = append(views, view)
	}
	return views
}

func (r *postgresAnalyticsRepository) Contacts() []*models.PropertyContact {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT `+contactColumns+` FROM property_contacts ORDER BY created_at ASC`)
	if err != nil {
		return []*models.PropertyContact{}
	}
	defer rows.Close()

	contacts := make([]*models.PropertyContact, 0)
	for rows.Next() {
		contact, err := scanContact(rows)
		if err != nil {
			return []*models.PropertyContact{}
		}
		contacts = append(contacts, contact)
	}
	return contacts
}
