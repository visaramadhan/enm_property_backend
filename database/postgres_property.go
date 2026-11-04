package database

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const propertyColumns = `id, partner_id, partner_name, slug, title, description, listing_type, ` +
	`property_type, category_id, address, city, province, district, latitude, longitude, price, ` +
	`price_period, land_area, building_area, bedrooms, bathrooms, facilities, units_total, ` +
	`units_available, status, views, contact_count, boost_active, boost_until, rating_count, ` +
	`rating_avg, favorite_count, published_at, deleted_at, created_at, updated_at`

const photoColumns = `id, property_id, url, storage_path, mime_type, size_bytes, width, height, ` +
	`sort_order, is_primary, created_at`

type postgresPropertyRepository struct{ pool pgxPool }

// scanProperty membaca satu baris properti dari hasil query.
func scanProperty(row pgx.Row) (*models.Property, error) {
	var property models.Property
	if err := row.Scan(
		&property.ID, &property.PartnerID, &property.PartnerName, &property.Slug,
		&property.Title, &property.Description, &property.ListingType, &property.PropertyType,
		&property.CategoryID, &property.Address, &property.City, &property.Province,
		&property.District, &property.Latitude, &property.Longitude, &property.Price,
		&property.PricePeriod, &property.LandArea, &property.BuildingArea, &property.Bedrooms,
		&property.Bathrooms, &property.Facilities, &property.UnitsTotal, &property.UnitsAvailable,
		&property.Status, &property.Views, &property.ContactCount, &property.BoostActive,
		&property.BoostUntil, &property.RatingCount, &property.RatingAvg, &property.FavoriteCount,
		&property.PublishedAt, &property.DeletedAt, &property.CreatedAt, &property.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &property, nil
}

// scanPhoto membaca satu baris foto properti dari hasil query.
func scanPhoto(row pgx.Row) (*models.PropertyPhoto, error) {
	var photo models.PropertyPhoto
	if err := row.Scan(
		&photo.ID, &photo.PropertyID, &photo.URL, &photo.StoragePath, &photo.MimeType,
		&photo.SizeBytes, &photo.Width, &photo.Height, &photo.SortOrder, &photo.IsPrimary,
		&photo.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &photo, nil
}

// loadPhotos memuat foto untuk sekumpulan ID properti, dikelompokkan per properti.
func (r *postgresPropertyRepository) loadPhotos(ctx context.Context, propertyIDs []string) (map[string][]models.PropertyPhoto, error) {
	grouped := make(map[string][]models.PropertyPhoto)
	if len(propertyIDs) == 0 {
		return grouped, nil
	}
	const query = `SELECT ` + photoColumns + ` FROM property_photos
		WHERE property_id = ANY($1::text[]) ORDER BY sort_order ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, query, propertyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		photo, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		grouped[photo.PropertyID] = append(grouped[photo.PropertyID], *photo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return grouped, nil
}

// attachPhotos menempelkan daftar foto ke properti (selalu slice tidak nil).
func attachPhotos(property *models.Property, grouped map[string][]models.PropertyPhoto) *models.Property {
	if property == nil {
		return nil
	}
	photos := grouped[property.ID]
	if photos == nil {
		photos = []models.PropertyPhoto{}
	}
	property.Photos = photos
	return property
}

// savePhotos menyimpan seluruh foto milik properti dengan pola upsert.
func (r *postgresPropertyRepository) savePhotos(ctx context.Context, propertyID string, photos []models.PropertyPhoto) error {
	for i := range photos {
		photo := photos[i]
		if photo.PropertyID == "" {
			photo.PropertyID = propertyID
		}
		if err := r.upsertPhoto(ctx, &photo); err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresPropertyRepository) upsertPhoto(ctx context.Context, photo *models.PropertyPhoto) error {
	const query = `INSERT INTO property_photos (` + photoColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			property_id = EXCLUDED.property_id, url = EXCLUDED.url,
			storage_path = EXCLUDED.storage_path, mime_type = EXCLUDED.mime_type,
			size_bytes = EXCLUDED.size_bytes, width = EXCLUDED.width, height = EXCLUDED.height,
			sort_order = EXCLUDED.sort_order, is_primary = EXCLUDED.is_primary,
			created_at = EXCLUDED.created_at`
	_, err := r.pool.Exec(ctx, query,
		photo.ID, photo.PropertyID, photo.URL, photo.StoragePath, photo.MimeType,
		photo.SizeBytes, photo.Width, photo.Height, photo.SortOrder, photo.IsPrimary,
		photo.CreatedAt,
	)
	return err
}

func (r *postgresPropertyRepository) Create(property *models.Property) error {
	ctx := context.Background()
	const query = `INSERT INTO properties (` + propertyColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36)`
	_, err := r.pool.Exec(ctx, query,
		property.ID, property.PartnerID, property.PartnerName, property.Slug, property.Title,
		property.Description, string(property.ListingType), string(property.PropertyType),
		property.CategoryID, property.Address, property.City, property.Province,
		property.District, property.Latitude, property.Longitude, property.Price,
		string(property.PricePeriod), property.LandArea, property.BuildingArea, property.Bedrooms,
		property.Bathrooms, property.Facilities, property.UnitsTotal, property.UnitsAvailable,
		string(property.Status), property.Views, property.ContactCount, property.BoostActive,
		property.BoostUntil, property.RatingCount, property.RatingAvg, property.FavoriteCount,
		property.PublishedAt, property.DeletedAt, property.CreatedAt, property.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return r.savePhotos(ctx, property.ID, property.Photos)
}

func (r *postgresPropertyRepository) Update(property *models.Property) error {
	ctx := context.Background()
	const query = `UPDATE properties SET
		partner_id = $2, partner_name = $3, slug = $4, title = $5, description = $6,
		listing_type = $7, property_type = $8, category_id = $9, address = $10, city = $11,
		province = $12, district = $13, latitude = $14, longitude = $15, price = $16,
		price_period = $17, land_area = $18, building_area = $19, bedrooms = $20,
		bathrooms = $21, facilities = $22, units_total = $23, units_available = $24,
		status = $25, views = $26, contact_count = $27, boost_active = $28, boost_until = $29,
		rating_count = $30, rating_avg = $31, favorite_count = $32, published_at = $33,
		deleted_at = $34, created_at = $35, updated_at = $36
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		property.ID, property.PartnerID, property.PartnerName, property.Slug, property.Title,
		property.Description, string(property.ListingType), string(property.PropertyType),
		property.CategoryID, property.Address, property.City, property.Province,
		property.District, property.Latitude, property.Longitude, property.Price,
		string(property.PricePeriod), property.LandArea, property.BuildingArea, property.Bedrooms,
		property.Bathrooms, property.Facilities, property.UnitsTotal, property.UnitsAvailable,
		string(property.Status), property.Views, property.ContactCount, property.BoostActive,
		property.BoostUntil, property.RatingCount, property.RatingAvg, property.FavoriteCount,
		property.PublishedAt, property.DeletedAt, property.CreatedAt, property.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return r.savePhotos(ctx, property.ID, property.Photos)
}

// Delete melakukan penghapusan lunak agar riwayat laporan tetap utuh.
func (r *postgresPropertyRepository) Delete(id string) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `UPDATE properties SET deleted_at = now(), updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresPropertyRepository) GetByID(id string) (*models.Property, error) {
	ctx := context.Background()
	property, err := scanProperty(r.pool.QueryRow(ctx, `SELECT `+propertyColumns+` FROM properties WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	grouped, err := r.loadPhotos(ctx, []string{property.ID})
	if err != nil {
		return nil, err
	}
	return attachPhotos(property, grouped), nil
}

func (r *postgresPropertyRepository) GetBySlug(slug string) (*models.Property, error) {
	ctx := context.Background()
	property, err := scanProperty(r.pool.QueryRow(ctx, `SELECT `+propertyColumns+` FROM properties WHERE lower(slug) = lower($1) LIMIT 1`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	grouped, err := r.loadPhotos(ctx, []string{property.ID})
	if err != nil {
		return nil, err
	}
	return attachPhotos(property, grouped), nil
}

func (r *postgresPropertyRepository) SlugExists(slug, excludeID string) bool {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM properties WHERE lower(slug) = lower($1) AND id <> $2`
	if err := r.pool.QueryRow(ctx, query, slug, excludeID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

// CountByPartner menghitung properti mitra tanpa menyertakan yang terhapus.
func (r *postgresPropertyRepository) CountByPartner(partnerID string, statuses ...models.PropertyStatus) int {
	ctx := context.Background()
	query := `SELECT count(*) FROM properties WHERE partner_id = $1 AND deleted_at IS NULL`
	args := []any{partnerID}
	if len(statuses) > 0 {
		values := make([]string, len(statuses))
		for i, status := range statuses {
			values[i] = string(status)
		}
		args = append(args, values)
		query += ` AND status = ANY($2::text[])`
	}
	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0
	}
	return count
}

// UpdateStatusByPartner memindahkan properti mitra dari satu status ke status lain.
func (r *postgresPropertyRepository) UpdateStatusByPartner(partnerID string, from, to models.PropertyStatus) int {
	ctx := context.Background()
	const query = `UPDATE properties SET status = $3, updated_at = now()
		WHERE partner_id = $1 AND deleted_at IS NULL AND status = $2`
	tag, err := r.pool.Exec(ctx, query, partnerID, string(from), string(to))
	if err != nil {
		return 0
	}
	return int(tag.RowsAffected())
}

func (r *postgresPropertyRepository) List(q ListQuery) ([]*models.Property, int, error) {
	ctx := context.Background()
	var w whereBuilder

	if q.Filter("include_deleted") != "true" {
		w.isNull("deleted_at")
	}
	if q.Filter("public") == "true" {
		w.eq("status", string(models.PropertyStatusActive))
	}
	w.eq("status", q.Filter("status"))
	w.eq("partner_id", q.Filter("partner_id"))
	w.eqFold("city", q.Filter("city"))
	w.eqFold("province", q.Filter("province"))
	w.eq("listing_type", q.Filter("listing_type"))
	w.eq("property_type", q.Filter("property_type"))
	w.eq("category_id", q.Filter("category_id"))
	if q.Filter("boost_active") == "true" {
		w.conds = append(w.conds, "boost_active = true")
	}
	if min := q.Filter("price_min"); min != "" {
		if value, err := strconv.ParseInt(min, 10, 64); err == nil {
			w.conds = append(w.conds, "price >= "+w.arg(value))
		}
	}
	if max := q.Filter("price_max"); max != "" {
		if value, err := strconv.ParseInt(max, 10, 64); err == nil {
			w.conds = append(w.conds, "price <= "+w.arg(value))
		}
	}
	w.searchOr([]string{"title", "city", "address", "partner_name"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "properties", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	sortField := q.SortBy
	desc := q.SortDesc
	if sortField == "" {
		sortField = "created_at"
		desc = true
	}
	order := orderClause(sortField, "created_at", desc, map[string]string{
		"price": "price",
		"views": "views",
		"title": "title",
	})

	query := "SELECT " + propertyColumns + " FROM properties" + whereSQL + " ORDER BY " + order + w.paginate(q)
	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	properties := make([]*models.Property, 0)
	ids := make([]string, 0)
	for rows.Next() {
		property, err := scanProperty(rows)
		if err != nil {
			return nil, 0, err
		}
		properties = append(properties, property)
		ids = append(ids, property.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	grouped, err := r.loadPhotos(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for _, property := range properties {
		attachPhotos(property, grouped)
	}
	return properties, total, nil
}

func (r *postgresPropertyRepository) CreatePhoto(photo *models.PropertyPhoto) error {
	return r.upsertPhoto(context.Background(), photo)
}

func (r *postgresPropertyRepository) UpdatePhoto(photo *models.PropertyPhoto) error {
	ctx := context.Background()
	const query = `UPDATE property_photos SET
		property_id = $2, url = $3, storage_path = $4, mime_type = $5, size_bytes = $6,
		width = $7, height = $8, sort_order = $9, is_primary = $10, created_at = $11
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		photo.ID, photo.PropertyID, photo.URL, photo.StoragePath, photo.MimeType,
		photo.SizeBytes, photo.Width, photo.Height, photo.SortOrder, photo.IsPrimary,
		photo.CreatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresPropertyRepository) DeletePhoto(id string) error {
	ctx := context.Background()
	tag, err := r.pool.Exec(ctx, `DELETE FROM property_photos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresPropertyRepository) GetPhoto(id string) (*models.PropertyPhoto, error) {
	ctx := context.Background()
	photo, err := scanPhoto(r.pool.QueryRow(ctx, `SELECT `+photoColumns+` FROM property_photos WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return photo, nil
}

func (r *postgresPropertyRepository) ListPhotos(propertyID string) ([]*models.PropertyPhoto, error) {
	ctx := context.Background()
	const query = `SELECT ` + photoColumns + ` FROM property_photos
		WHERE property_id = $1 ORDER BY sort_order ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := make([]*models.PropertyPhoto, 0)
	for rows.Next() {
		photo, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return photos, nil
}
