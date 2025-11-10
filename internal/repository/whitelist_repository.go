package repository

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"strings"
)

type WhitelistRepository interface {
	// whitelists 本体
	Add(ctx context.Context, w models.Whitelist) error
	Remove(ctx context.Context, platform models.WhitelistPlatform, userID string) error
	Exists(ctx context.Context, platform models.WhitelistPlatform, userID string) (bool, error)
	List(ctx context.Context, platform models.WhitelistPlatform, limit, offset int) ([]models.Whitelist, error)

	// 1ユーザの行取得 / なければ作成
	Get(ctx context.Context, platform models.WhitelistPlatform, userID string) (*models.Whitelist, error)
	GetOrCreate(ctx context.Context, platform models.WhitelistPlatform, userID, note string) (*models.Whitelist, error)

	// whitelist_items（VRC名ひも付け）
	AddItemIfNotExists(ctx context.Context, whitelistID uint64, vrcName string) (bool, error)
	ListItemsByWhitelistID(ctx context.Context, whitelistID uint64) ([]string, error)
}

type whitelistRepository struct {
	db *sql.DB
}

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

//
// whitelists 本体
//

func (r *whitelistRepository) Add(ctx context.Context, w models.Whitelist) error {
	// 既にあれば note を更新、なければ作成。
	// （サービス側で note をどう使うかに応じてここを変える余地あり）
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO whitelists (platform, user_id, note)
         VALUES (?, ?, ?)
         ON DUPLICATE KEY UPDATE note = VALUES(note)`,
		string(w.Platform), w.UserID, w.Note,
	)
	return err
}

func (r *whitelistRepository) Remove(ctx context.Context, platform models.WhitelistPlatform, userID string) error {
	// ON DELETE CASCADE により whitelist_items もまとめて消える想定
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM whitelists
		  WHERE platform = ? AND user_id = ?`,
		string(platform), userID,
	)
	return err
}

func (r *whitelistRepository) Exists(ctx context.Context, platform models.WhitelistPlatform, userID string) (bool, error) {
	const q = `
		SELECT 1
		FROM whitelists
		WHERE platform = ? AND user_id = ?
		LIMIT 1`
	var dummy int
	err := r.db.QueryRowContext(ctx, q, string(platform), userID).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *whitelistRepository) List(ctx context.Context, platform models.WhitelistPlatform, limit, offset int) ([]models.Whitelist, error) {
	const q = `
		SELECT id, platform, user_id, note, created_at
		FROM whitelists
		WHERE platform = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, q, string(platform), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Whitelist
	for rows.Next() {
		var w models.Whitelist
		var p string
		var createdAt any // 使わないので捨てる

		if err := rows.Scan(&w.ID, &p, &w.UserID, &w.Note, &createdAt); err != nil {
			return nil, err
		}
		w.Platform = models.WhitelistPlatform(p)
		result = append(result, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *whitelistRepository) Get(
	ctx context.Context,
	platform models.WhitelistPlatform,
	userID string,
) (*models.Whitelist, error) {
	const q = `
		SELECT id, platform, user_id, note
		FROM whitelists
		WHERE platform = ? AND user_id = ?
		LIMIT 1`
	var w models.Whitelist
	var p string

	err := r.db.QueryRowContext(ctx, q, string(platform), userID).
		Scan(&w.ID, &p, &w.UserID, &w.Note)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.Platform = models.WhitelistPlatform(p)
	return &w, nil
}

func (r *whitelistRepository) GetOrCreate(
	ctx context.Context,
	platform models.WhitelistPlatform,
	userID string,
	note string,
) (*models.Whitelist, error) {
	// まず既存確認
	wl, err := r.Get(ctx, platform, userID)
	if err != nil {
		return nil, err
	}
	if wl != nil {
		return wl, nil
	}

	// なければ作成。
	// 競合を考慮して INSERT ... ON DUP を使い、その後 SELECT。
	const insertQ = `
		INSERT INTO whitelists (platform, user_id, note)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE note = VALUES(note)
		`
	_, err = r.db.ExecContext(ctx, insertQ, string(platform), userID, note)
	if err != nil {
		return nil, err
	}

	// 必ず存在する状態になっているので再取得
	return r.Get(ctx, platform, userID)
}

//
// whitelist_items（VRC名ひもづけ）
//

func (r *whitelistRepository) AddItemIfNotExists(
	ctx context.Context,
	whitelistID uint64,
	vrcName string,
) (bool, error) {
	vrcName = strings.TrimSpace(vrcName)
	if vrcName == "" {
		return false, nil
	}

	// vrc_name_norm は GENERATED COLUMN。
	// UNIQUE(whitelist_id, vrc_name_norm) に任せてUPSERTする。
	const insertQ = `
		INSERT INTO whitelist_items (whitelist_id, vrc_name)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE id = id
		`
	res, err := r.db.ExecContext(ctx, insertQ, whitelistID, vrcName)
	if err != nil {
		return false, err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	// 新規追加: true / 既に存在: false
	return aff > 0, nil
}

func (r *whitelistRepository) ListItemsByWhitelistID(
	ctx context.Context,
	whitelistID uint64,
) ([]string, error) {
	const q = `
		SELECT vrc_name
		FROM whitelist_items
		WHERE whitelist_id = ?
		ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, q, whitelistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
