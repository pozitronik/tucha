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
func (r *UserRepository) Upsert(email, password string, isAdmin bool, quotaBytes int64) (int64, error) {
	now := time.Now().Unix()

	res, err := r.db.Exec(
		`INSERT INTO users (email, password, is_admin, quota_bytes, created) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(email) DO UPDATE SET password = excluded.password, is_admin = excluded.is_admin`,
		email, password, boolToInt(isAdmin), quotaBytes, now,
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

// Create inserts a new user. Returns the user ID.
func (r *UserRepository) Create(user *entity.User) (int64, error) {
	now := time.Now().Unix()

	res, err := r.db.Exec(
		`INSERT INTO users (email, password, is_admin, quota_bytes, created) VALUES (?, ?, ?, ?, ?)`,
		user.Email, user.Password, boolToInt(user.IsAdmin), user.QuotaBytes, now,
	)
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting user id: %w", err)
	}

	return id, nil
}

// GetByID retrieves a user by ID. Returns nil, nil if not found.
func (r *UserRepository) GetByID(id int64) (*entity.User, error) {
	u := &entity.User{}
	var isAdmin int
	err := r.db.QueryRow(
		"SELECT id, email, password, is_admin, quota_bytes, created FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Email, &u.Password, &isAdmin, &u.QuotaBytes, &u.Created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	u.IsAdmin = isAdmin != 0
	return u, nil
}

// GetByEmail retrieves a user by email. Returns nil, nil if not found.
func (r *UserRepository) GetByEmail(email string) (*entity.User, error) {
	u := &entity.User{}
	var isAdmin int
	err := r.db.QueryRow(
		"SELECT id, email, password, is_admin, quota_bytes, created FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &isAdmin, &u.QuotaBytes, &u.Created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	u.IsAdmin = isAdmin != 0
	return u, nil
}

// List returns all users.
func (r *UserRepository) List() ([]entity.User, error) {
	rows, err := r.db.Query("SELECT id, email, password, is_admin, quota_bytes, created FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		var isAdmin int
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &isAdmin, &u.QuotaBytes, &u.Created); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		u.IsAdmin = isAdmin != 0
		users = append(users, u)
	}
	return users, rows.Err()
}

// Update modifies an existing user's fields.
func (r *UserRepository) Update(user *entity.User) error {
	res, err := r.db.Exec(
		`UPDATE users SET email = ?, password = ?, is_admin = ?, quota_bytes = ? WHERE id = ?`,
		user.Email, user.Password, boolToInt(user.IsAdmin), user.QuotaBytes, user.ID,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("user %d not found", user.ID)
	}
	return nil
}

// Delete removes a user by ID. Tokens cascade via FK constraint.
func (r *UserRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("user %d not found", id)
	}
	return nil
}

// boolToInt converts a boolean to an integer for SQLite storage.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
