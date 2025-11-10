package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrAlreadyExists      = errors.New("already exists")            // VRC ID が他ユーザに紐付いてる
	ErrNoExactMatch       = errors.New("no exact match user found") // 完全一致0件
	ErrMultipleExactMatch = errors.New("multiple exact matches")    // 完全一致が複数
)

type WhitelistService interface {
	// Discord ID で許可されてるか
	IsAllowedByDiscord(ctx context.Context, discordID string) (bool, error)
	// VRC userId で許可されてるか（ワールド側チェック用など）
	IsAllowedByVRCUserID(ctx context.Context, vrcUserID string) (bool, error)

	// Discord ユーザに VRC displayName を指定して紐付け登録
	// 完全一致 1件のみ成功。既存紐付けあれば上書き。
	RegisterDiscordVRC(ctx context.Context, discordID, vrcDisplayName string) (created bool, err error)

	// Discord に紐付いている情報を取得
	GetDiscordVRC(ctx context.Context, discordID string) (*models.WhitelistUser, error)

	// Discord ユーザの紐付け解除
	UnlinkDiscord(ctx context.Context, discordID string) error
}

type whitelistService struct {
	repo   repository.WhitelistRepository
	vrchat VRChatClient
}

func NewWhitelistService(r repository.WhitelistRepository, v VRChatClient) WhitelistService {
	return &whitelistService{
		repo:   r,
		vrchat: v,
	}
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

func (s *whitelistService) RegisterDiscordVRC(
	ctx context.Context,
	discordID string,
	vrcDisplayName string,
) (bool, error) {
	discordID = strings.TrimSpace(discordID)
	vrcDisplayName = strings.TrimSpace(vrcDisplayName)

	if discordID == "" || vrcDisplayName == "" {
		return false, ErrInvalidArgument
	}

	// 1. VRChat API で検索
	users, err := s.vrchat.SearchUsers(ctx, vrcDisplayName, 100)
	if err != nil {
		return false, err
	}

	// 2. displayName 完全一致のみ抽出
	var matches []VRChatUser
	for _, u := range users {
		if u.DisplayName == vrcDisplayName {
			matches = append(matches, u)
		}
	}

	if len(matches) == 0 {
		return false, ErrNoExactMatch
	}
	if len(matches) > 1 {
		return false, ErrMultipleExactMatch
	}

	match := matches[0]

	// 3. その VRC userId が既に他ユーザに使われていないかチェック
	existingByVRC, err := s.repo.GetByVRCUserID(ctx, match.ID)
	if err != nil {
		return false, err
	}
	if existingByVRC != nil && existingByVRC.DiscordUserID != discordID {
		// 他人がこの VRC アカウントを既に使っている
		return false, ErrAlreadyExists
	}

	// 4. 自分の既存レコードを確認（新規か更新か判定したいだけ）
	existingByDiscord, err := s.repo.GetByDiscordID(ctx, discordID)
	if err != nil {
		return false, err
	}

	u := &models.WhitelistUser{
		DiscordUserID:  discordID,
		VRCUserID:      match.ID,
		VRCDisplayName: match.DisplayName,
		Note:           "",
	}
	if err := s.repo.Upsert(ctx, u); err != nil {
		return false, err
	}

	created := (existingByDiscord == nil)
	return created, nil
}

func (s *whitelistService) GetDiscordVRC(ctx context.Context, discordID string) (*models.WhitelistUser, error) {
	if discordID == "" {
		return nil, ErrInvalidArgument
	}
	return s.repo.GetByDiscordID(ctx, discordID)
}

func (s *whitelistService) UnlinkDiscord(ctx context.Context, discordID string) error {
	if discordID == "" {
		return ErrInvalidArgument
	}
	return s.repo.RemoveByDiscordID(ctx, discordID)
}
