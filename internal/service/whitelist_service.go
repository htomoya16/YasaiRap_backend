package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"context"
	"errors"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExists   = errors.New("already exists")
)

type WhitelistService interface {
	// 既存
	Add(ctx context.Context, platform models.WhitelistPlatform, userID, note string) error
	Remove(ctx context.Context, platform models.WhitelistPlatform, userID string) error
	IsAllowed(ctx context.Context, platform models.WhitelistPlatform, userID string) (bool, error)

	// DiscordユーザにVRC名を紐づけて登録
	RegisterDiscordVRC(ctx context.Context, discordID, vrcName string) (created bool, err error)
	GetDiscordVRCNames(ctx context.Context, discordID string) ([]string, error)
}

type whitelistService struct {
	repo repository.WhitelistRepository
}

func NewWhitelistService(r repository.WhitelistRepository) WhitelistService {
	return &whitelistService{repo: r}
}

func (s *whitelistService) Add(ctx context.Context, platform models.WhitelistPlatform, userID, note string) error {
	if userID == "" {
		return ErrInvalidArgument
	}
	return s.repo.Add(ctx, models.Whitelist{
		Platform: platform,
		UserID:   userID,
		Note:     note,
	})
}

func (s *whitelistService) Remove(ctx context.Context, platform models.WhitelistPlatform, userID string) error {
	if userID == "" {
		return ErrInvalidArgument
	}
	return s.repo.Remove(ctx, platform, userID)
}

func (s *whitelistService) IsAllowed(ctx context.Context, platform models.WhitelistPlatform, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	return s.repo.Exists(ctx, platform, userID)
}

// Discord ID と VRC名を紐づけて whitelist_items に登録
func (s *whitelistService) RegisterDiscordVRC(
	ctx context.Context,
	discordID string,
	vrcName string,
) (bool, error) {
	if discordID == "" || vrcName == "" {
		return false, ErrInvalidArgument
	}

	// 親レコード取得 or 作成
	wl, err := s.repo.GetOrCreate(ctx, models.WhitelistPlatformDiscord, discordID, "")
	if err != nil {
		return false, err
	}

	// 子レコード追加（重複時は何も起きない）
	created, err := s.repo.AddItemIfNotExists(ctx, wl.ID, vrcName)

	if err != nil {
		return false, err
	}
	if !created {
		return false, ErrAlreadyExists
	}
	return true, nil
}

func (s *whitelistService) GetDiscordVRCNames(ctx context.Context, discordID string) ([]string, error) {
	if discordID == "" {
		return nil, nil
	}

	wl, err := s.repo.Get(ctx, models.WhitelistPlatformDiscord, discordID)
	if err != nil {
		return nil, err
	}
	if wl == nil {
		return []string{}, nil
	}

	return s.repo.ListItemsByWhitelistID(ctx, wl.ID)
}
