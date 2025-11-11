package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExists   = errors.New("already exists") // 他人がそのVRC IDを使用
)

type WhitelistService interface {
	RegisterDiscordVRC(ctx context.Context, discordID, vrcDisplayName string) (created bool, err error)
	GetDiscordVRC(ctx context.Context, discordID string) (*models.WhitelistUser, error)
	IsAllowedByDiscord(ctx context.Context, discordID string) (bool, error)
	IsAllowedByVRCUserID(ctx context.Context, vrcUserID string) (bool, error)
	RemoveDiscord(ctx context.Context, discordID string) error
}

type whitelistService struct {
	repo   repository.WhitelistRepository
	vrchat VRChatClient
}

func NewWhitelistService(repo repository.WhitelistRepository, vrchat VRChatClient) WhitelistService {
	return &whitelistService{
		repo:   repo,
		vrchat: vrchat,
	}
}

func (s *whitelistService) RegisterDiscordVRC(
	ctx context.Context,
	discordID string,
	vrcDisplayName string,
) (bool, error) {
	// 空白削除
	discordID = strings.TrimSpace(discordID)
	vrcDisplayName = strings.TrimSpace(vrcDisplayName)
	// どっちか空
	if discordID == "" || vrcDisplayName == "" {
		return false, ErrInvalidArgument
	}

	// 1. VRChat APIで完全一致検索
	user, err := s.vrchat.SearchExactUserByDisplayName(ctx, vrcDisplayName)
	if err != nil {
		// ErrNoExactMatch / ErrMultipleExactMatch はそのまま上に返してハンドラー側で文言出す
		if errors.Is(err, ErrNoExactMatch) || errors.Is(err, ErrMultipleExactMatch) {
			return false, err
		}
		return false, err
	}

	// 2. その VRC userID が他人に使われていないか確認
	existingByVRC, err := s.repo.GetByVRCUserID(ctx, user.ID)
	// 使われていたらエラー
	if err != nil {
		return false, err
	}
	// 他人のVRCアカウントは使っちゃダメ
	if existingByVRC != nil && existingByVRC.DiscordUserID != discordID {
		return false, ErrAlreadyExists
	}

	// 3. Discordユーザの既存レコード確認
	existingByDiscord, err := s.repo.GetByDiscordID(ctx, discordID)
	// このDiscordユーザーは、既に何かVRCアカウントと紐づいてたらエラー
	if err != nil {
		return false, err
	}

	u := &models.WhitelistUser{
		DiscordUserID:  discordID,
		VRCUserID:      user.ID,
		VRCDisplayName: user.DisplayName,
		VRCAvatarURL:   user.CurrentAvatarImageURL,
		Note:           "",
	}

	// 4. Upsert で (discordID, userID) を保存
	if err := s.repo.Upsert(ctx, u); err != nil {
		return false, err
	}

	// created = true → 新規, false → 新規ではなく更新
	created := (existingByDiscord == nil)
	return created, nil
}

func (s *whitelistService) GetDiscordVRC(ctx context.Context, discordID string) (*models.WhitelistUser, error) {
	if discordID == "" {
		return nil, ErrInvalidArgument
	}
	return s.repo.GetByDiscordID(ctx, discordID)
}

func (s *whitelistService) IsAllowedByDiscord(ctx context.Context, discordID string) (bool, error) {
	if discordID == "" {
		return false, ErrInvalidArgument
	}
	return s.repo.ExistsByDiscordID(ctx, discordID)
}

func (s *whitelistService) IsAllowedByVRCUserID(ctx context.Context, vrcUserID string) (bool, error) {
	if vrcUserID == "" {
		return false, ErrInvalidArgument
	}
	return s.repo.ExistsByVRCUserID(ctx, vrcUserID)
}

func (s *whitelistService) RemoveDiscord(ctx context.Context, discordID string) error {
	if discordID == "" {
		return ErrInvalidArgument
	}
	return s.repo.RemoveByDiscordID(ctx, discordID)
}
