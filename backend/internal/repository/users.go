package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/wahyu241205/SignalArc/backend/internal/database"
)

type User struct {
	ID          string         `json:"id"`
	ExternalID  sql.NullString `json:"external_id"`
	Email       sql.NullString `json:"email"`
	DisplayName sql.NullString `json:"display_name"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type UsersRepository struct {
	db *database.DB
}

func NewUsersRepository(db *database.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) GetUserByID(ctx context.Context, id string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id::text, external_id, email, display_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.ExternalID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func (r *UsersRepository) ListUsers(ctx context.Context, limit int) ([]User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, external_id, email, display_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.ExternalID,
			&user.Email,
			&user.DisplayName,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
