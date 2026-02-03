package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"tucha/internal/domain/entity"
)

// UserRepository implements repository.UserRepository using SQLite.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a UserRepository from the given database connection.
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db.Conn()}
}

// Upsert creates or updates a user by email. Returns the user ID.
func (r *UserRepository) Upsert(email, password string) (int64, error) {
	now := time.Now().Unix()

	res, err := r.db.Exec(
		`INSERT INTO users (email, password, created) VALUES (?, ?, ?)
		 ON CONFLICT(email) DO UPDATE SET password = excluded.password`,
		email, password, now,
	)
	if err != nil {
		return 0, fmt.Errorf("upserting user: %w", err)
	}

	userID, err := res.LastInsertId()
	if err != nil || userID == 0 {
		row := r.db.QueryRow("SELECT id FROM users WHERE email = ?", email)
		if err := row.Scan(&userID); err != nil {
			return 0, fmt.Errorf("looking up user: %w", err)
		}
	}

	return userID, nil
}

// GetByEmail retrieves a user by email. Returns nil, nil if not found.
func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	u := &entity.User{}
	err := r.db.QueryRow(
		"SELECT id, email, password, created FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return u, nil
}
