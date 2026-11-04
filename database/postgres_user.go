package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"enmproperty/backend/models"
)

const userColumns = `id, firebase_uid, email, name, phone, avatar_url, role, status, ` +
	`email_verified, last_login_at, created_at, updated_at`

type postgresUserRepository struct{ pool pgxPool }

// scanUser membaca satu baris pengguna dari hasil query.
func scanUser(row pgx.Row) (*models.User, error) {
	var user models.User
	if err := row.Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.Name, &user.Phone, &user.AvatarURL,
		&user.Role, &user.Status, &user.EmailVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) Create(user *models.User) error {
	ctx := context.Background()

	if user.FirebaseUID != "" {
		var count int
		if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE firebase_uid = $1`, user.FirebaseUID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return errors.New("firebase uid sudah terdaftar")
		}
	}
	if user.Email != "" {
		var count int
		if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email = $1`, user.Email).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return errors.New("email sudah terdaftar")
		}
	}

	const query = `INSERT INTO users (` + userColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.FirebaseUID, user.Email, user.Name, user.Phone, user.AvatarURL,
		string(user.Role), string(user.Status), user.EmailVerified, user.LastLoginAt,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) Update(user *models.User) error {
	ctx := context.Background()
	const query = `UPDATE users SET
		firebase_uid = $2, email = $3, name = $4, phone = $5, avatar_url = $6,
		role = $7, status = $8, email_verified = $9, last_login_at = $10,
		created_at = $11, updated_at = $12
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		user.ID, user.FirebaseUID, user.Email, user.Name, user.Phone, user.AvatarURL,
		string(user.Role), string(user.Status), user.EmailVerified, user.LastLoginAt,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresUserRepository) GetByID(id string) (*models.User, error) {
	ctx := context.Background()
	user, err := scanUser(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) GetByFirebaseUID(uid string) (*models.User, error) {
	ctx := context.Background()
	user, err := scanUser(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE firebase_uid = $1 LIMIT 1`, uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) GetByEmail(email string) (*models.User, error) {
	ctx := context.Background()
	user, err := scanUser(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1) LIMIT 1`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) List(q ListQuery) ([]*models.User, int, error) {
	ctx := context.Background()
	var w whereBuilder

	w.eq("role", q.Filter("role"))
	w.eq("status", q.Filter("status"))
	w.searchOr([]string{"name", "email"}, q.Search)

	whereSQL := w.where()
	total, err := countRows(ctx, r.pool, "users", whereSQL, w.args)
	if err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if q.SortBy == "name" {
		order = "name ASC"
	}
	query := "SELECT " + userColumns + " FROM users" + whereSQL + " ORDER BY " + order + w.paginate(q)

	rows, err := r.pool.Query(ctx, query, w.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
