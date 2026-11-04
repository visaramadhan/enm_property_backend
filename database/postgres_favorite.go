package database

import (
	"context"

	"enmproperty/backend/models"
)

const favoriteColumns = `id, user_id, property_id, created_at`

type postgresFavoriteRepository struct{ pool pgxPool }

// scanFavorite membaca satu baris favorit dari hasil query.
func scanFavorite(row interface{ Scan(...any) error }) (*models.Favorite, error) {
	var favorite models.Favorite
	if err := row.Scan(&favorite.ID, &favorite.UserID, &favorite.PropertyID, &favorite.CreatedAt); err != nil {
		return nil, err
	}
	return &favorite, nil
}

// Add menyimpan favorit dan mengabaikan duplikasi user-properti.
func (r *postgresFavoriteRepository) Add(favorite *models.Favorite) error {
	ctx := context.Background()
	const query = `INSERT INTO favorites (` + favoriteColumns + `)
		VALUES ($1,$2,$3,$4) ON CONFLICT (user_id, property_id) DO NOTHING`
	_, err := r.pool.Exec(ctx, query,
		favorite.ID, favorite.UserID, favorite.PropertyID, favorite.CreatedAt,
	)
	return err
}

func (r *postgresFavoriteRepository) Remove(userID, propertyID string) error {
	ctx := context.Background()
	const query = `DELETE FROM favorites WHERE user_id = $1 AND property_id = $2`
	tag, err := r.pool.Exec(ctx, query, userID, propertyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresFavoriteRepository) Exists(userID, propertyID string) bool {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM favorites WHERE user_id = $1 AND property_id = $2`
	if err := r.pool.QueryRow(ctx, query, userID, propertyID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (r *postgresFavoriteRepository) ListByUser(userID string) ([]*models.Favorite, error) {
	ctx := context.Background()
	const query = `SELECT ` + favoriteColumns + ` FROM favorites WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favorites := make([]*models.Favorite, 0)
	for rows.Next() {
		favorite, err := scanFavorite(rows)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, favorite)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return favorites, nil
}

func (r *postgresFavoriteRepository) CountByProperty(propertyID string) int {
	ctx := context.Background()
	var count int
	const query = `SELECT count(*) FROM favorites WHERE property_id = $1`
	if err := r.pool.QueryRow(ctx, query, propertyID).Scan(&count); err != nil {
		return 0
	}
	return count
}
