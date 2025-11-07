package repository

import (
	"backend/internal/models"
	"context"
	"database/sql"
)

type WhitelistRepository interface {
	Add(ctx context.Context, w models.DiscordWhitelist) error
	Remove(ctx context.Context, platform, userID string) error
	Exists(ctx context.Context, platform, userID string) (bool, error)
	List(ctx context.Context, platform string, limit, offset int) ([]models.DiscordWhitelist, error)
}

type whitelistRepository struct{ db *sql.DB }

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

func (r *whitelistRepository) Add(ctx context.Context, w models.DiscordWhitelist) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO whitelist (platform, user_id, note, created_at)
		 VALUES (?, ?, ?, UNIX_TIMESTAMP())`,
		w.Platform, w.UserID, w.Note,
	)
	return err
}

func (r *whitelistRepository) Remove(ctx context.Context, platform, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM whitelist WHERE platform=? AND user_id=?`, platform, userID)
	return err
}

func (r *whitelistRepository) Exists(ctx context.Context, platform, userID string) (bool, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM whitelist WHERE platform=? AND user_id=? LIMIT 1`, platform, userID)
	var x int
	if err := row.Scan(&x); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *whitelistRepository) List(ctx context.Context, platform string, limit, offset int) ([]models.DiscordWhitelist, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, platform, user_id, note, created_at
		   FROM whitelist
		  WHERE platform=?
		  ORDER BY id DESC
		  LIMIT ? OFFSET ?`, platform, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DiscordWhitelist
	for rows.Next() {
		var w models.DiscordWhitelist
		if err := rows.Scan(&w.ID, &w.Platform, &w.UserID, &w.Note, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
