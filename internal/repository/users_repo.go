package repository

import (
	"context"

	"github.com/alloy/diffinder/internal/model"
	"github.com/google/uuid"
)

type UsersRepo struct{ db *DB }

func NewUsersRepo(db *DB) *UsersRepo { return &UsersRepo{db: db} }

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, username, email, password_hash, role, created_at
		FROM users WHERE email = $1`
	var u model.User
	err := r.db.Pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const q = `
		SELECT id, username, email, password_hash, role, created_at
		FROM users WHERE id = $1`
	var u model.User
	err := r.db.Pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersRepo) Create(ctx context.Context, u *model.User) error {
	const q = `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return r.db.Pool.QueryRow(ctx, q,
		u.Username, u.Email, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *UsersRepo) List(ctx context.Context, limit, offset int) ([]model.User, int, error) {
	const q = `
		SELECT id, username, email, role, created_at,
		       COUNT(*) OVER()
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.User
	var total int
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}
