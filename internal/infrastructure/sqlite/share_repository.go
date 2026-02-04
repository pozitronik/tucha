package sqlite

import (
	"database/sql"
	"fmt"

	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// ShareRepository implements repository.ShareRepository using SQLite.
type ShareRepository struct {
	db *sql.DB
}

// NewShareRepository creates a ShareRepository from the given database connection.
func NewShareRepository(db *DB) *ShareRepository {
	return &ShareRepository{db: db.Conn()}
}

// shareColumns is the standard column list for share queries.
const shareColumns = `id, owner_id, home, invited_email, access, status, invite_token, mount_home, mount_user_id, created`

// Create inserts a new share invitation.
func (r *ShareRepository) Create(share *entity.Share) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO shares (owner_id, home, invited_email, access, status, invite_token)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		share.OwnerID, share.Home.String(), share.InvitedEmail,
		share.Access.String(), share.Status.String(), share.InviteToken,
	)
	if err != nil {
		return 0, fmt.Errorf("creating share: %w", err)
	}
	return res.LastInsertId()
}

// GetByOwnerPathEmail finds a share by owner, path, and invited email.
func (r *ShareRepository) GetByOwnerPathEmail(ownerID int64, home vo.CloudPath, email string) (*entity.Share, error) {
	row := r.db.QueryRow(
		`SELECT `+shareColumns+` FROM shares WHERE owner_id = ? AND home = ? AND invited_email = ?`,
		ownerID, home.String(), email,
	)
	return scanShare(row)
}

// GetByInviteToken finds a share by its invite token.
func (r *ShareRepository) GetByInviteToken(token string) (*entity.Share, error) {
	row := r.db.QueryRow(
		`SELECT `+shareColumns+` FROM shares WHERE invite_token = ?`,
		token,
	)
	return scanShare(row)
}

// Delete removes a share by ID.
func (r *ShareRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM shares WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting share: %w", err)
	}
	return nil
}

// ListByOwnerPath returns all shares for a given owner and path.
func (r *ShareRepository) ListByOwnerPath(ownerID int64, home vo.CloudPath) ([]entity.Share, error) {
	rows, err := r.db.Query(
		`SELECT `+shareColumns+` FROM shares WHERE owner_id = ? AND home = ? ORDER BY created DESC`,
		ownerID, home.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("listing shares: %w", err)
	}
	defer rows.Close()
	return scanShares(rows)
}

// ListIncoming returns all pending shares where the invited email matches.
func (r *ShareRepository) ListIncoming(email string) ([]entity.Share, error) {
	rows, err := r.db.Query(
		`SELECT `+shareColumns+` FROM shares WHERE invited_email = ? AND status = 'pending' ORDER BY created DESC`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("listing incoming shares: %w", err)
	}
	defer rows.Close()
	return scanShares(rows)
}

// Accept transitions a share to "accepted" and records the mount details.
func (r *ShareRepository) Accept(inviteToken string, mountUserID int64, mountHome string) error {
	_, err := r.db.Exec(
		`UPDATE shares SET status = 'accepted', mount_user_id = ?, mount_home = ? WHERE invite_token = ?`,
		mountUserID, mountHome, inviteToken,
	)
	if err != nil {
		return fmt.Errorf("accepting share: %w", err)
	}
	return nil
}

// Reject transitions a share to "rejected".
func (r *ShareRepository) Reject(inviteToken string) error {
	_, err := r.db.Exec(
		`UPDATE shares SET status = 'rejected' WHERE invite_token = ?`,
		inviteToken,
	)
	if err != nil {
		return fmt.Errorf("rejecting share: %w", err)
	}
	return nil
}

// Unmount clears mount details and transitions the share back to "pending".
func (r *ShareRepository) Unmount(userID int64, mountHome string) (*entity.Share, error) {
	// Fetch the share first so we can return it.
	share, err := r.GetByMountPath(userID, mountHome)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, nil
	}

	_, err = r.db.Exec(
		`UPDATE shares SET status = 'pending', mount_user_id = NULL, mount_home = NULL WHERE id = ?`,
		share.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("unmounting share: %w", err)
	}

	return share, nil
}

// GetByMountPath finds a share by mount user and mount path.
func (r *ShareRepository) GetByMountPath(userID int64, mountHome string) (*entity.Share, error) {
	row := r.db.QueryRow(
		`SELECT `+shareColumns+` FROM shares WHERE mount_user_id = ? AND mount_home = ? AND status = 'accepted'`,
		userID, mountHome,
	)
	return scanShare(row)
}

// scanShare scans a single share row into an entity.Share.
func scanShare(s interface{ Scan(...any) error }) (*entity.Share, error) {
	var (
		share       entity.Share
		home        string
		access      string
		status      string
		mountHome   sql.NullString
		mountUserID sql.NullInt64
	)

	err := s.Scan(
		&share.ID, &share.OwnerID, &home, &share.InvitedEmail,
		&access, &status, &share.InviteToken,
		&mountHome, &mountUserID, &share.Created,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning share: %w", err)
	}

	share.Home = vo.NewCloudPath(home)
	share.Access, _ = vo.ParseAccessLevel(access)
	share.Status, _ = vo.ParseShareStatus(status)
	if mountHome.Valid {
		share.MountHome = mountHome.String
	}
	if mountUserID.Valid {
		share.MountUserID = &mountUserID.Int64
	}

	return &share, nil
}

// scanShares scans multiple share rows into a slice.
func scanShares(rows *sql.Rows) ([]entity.Share, error) {
	var shares []entity.Share
	for rows.Next() {
		share, err := scanShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, *share)
	}
	return shares, rows.Err()
}
