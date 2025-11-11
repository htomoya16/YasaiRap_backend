package repository

import (
	"backend/internal/models"
	"context"
	"database/sql"
)

type WhitelistRepository interface {
	Upsert(ctx context.Context, u *models.WhitelistUser) error
	GetByDiscordID(ctx context.Context, discordID string) (*models.WhitelistUser, error)
	GetByVRCUserID(ctx context.Context, vrcUserID string) (*models.WhitelistUser, error)
	ExistsByDiscordID(ctx context.Context, discordID string) (bool, error)
	ExistsByVRCUserID(ctx context.Context, vrcUserID string) (bool, error)
	RemoveByDiscordID(ctx context.Context, discordID string) error
}

type whitelistRepository struct {
	db *sql.DB
}

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

func (r *whitelistRepository) Upsert(ctx context.Context, u *models.WhitelistUser) error {
	// ON DUPLICATE KEY UPDATE レコードを挿入または重複があった場合は更新
	const q = `
		INSERT INTO whitelist_users (
		discord_user_id,
		vrc_user_id,
		vrc_display_name,
		note
		) VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		vrc_user_id      = VALUES(vrc_user_id),
		vrc_display_name = VALUES(vrc_display_name),
		note             = VALUES(note),
		updated_at       = CURRENT_TIMESTAMP(6);
		`
	_, err := r.db.ExecContext(ctx, q,
		u.DiscordUserID,
		u.VRCUserID,
		u.VRCDisplayName,
		u.Note,
	)
	return err
}

func (r *whitelistRepository) GetByDiscordID(ctx context.Context, discordID string) (*models.WhitelistUser, error) {
	const q = `
		SELECT id, discord_user_id, vrc_user_id, vrc_display_name, note, created_at, updated_at
		FROM whitelist_users
		WHERE discord_user_id = ?
		LIMIT 1;
		`
	row := r.db.QueryRowContext(ctx, q, discordID)

	var u models.WhitelistUser
	if err := row.Scan(
		&u.ID,
		&u.DiscordUserID,
		&u.VRCUserID,
		&u.VRCDisplayName,
		&u.Note,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *whitelistRepository) GetByVRCUserID(ctx context.Context, vrcUserID string) (*models.WhitelistUser, error) {
	const q = `
		SELECT id, discord_user_id, vrc_user_id, vrc_display_name, note, created_at, updated_at
		FROM whitelist_users
		WHERE vrc_user_id = ?
		LIMIT 1;
		`
	row := r.db.QueryRowContext(ctx, q, vrcUserID)

	var u models.WhitelistUser
	if err := row.Scan(
		&u.ID,
		&u.DiscordUserID,
		&u.VRCUserID,
		&u.VRCDisplayName,
		&u.Note,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *whitelistRepository) ExistsByDiscordID(ctx context.Context, discordID string) (bool, error) {
	const q = `SELECT 1 FROM whitelist_users WHERE discord_user_id = ? LIMIT 1`
	var x int
	err := r.db.QueryRowContext(ctx, q, discordID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *whitelistRepository) ExistsByVRCUserID(ctx context.Context, vrcUserID string) (bool, error) {
	const q = `SELECT 1 FROM whitelist_users WHERE vrc_user_id = ? LIMIT 1`
	var x int
	err := r.db.QueryRowContext(ctx, q, vrcUserID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *whitelistRepository) RemoveByDiscordID(ctx context.Context, discordID string) error {
	const q = `DELETE FROM whitelist_users WHERE discord_user_id = ?`
	_, err := r.db.ExecContext(ctx, q, discordID)
	return err
}
