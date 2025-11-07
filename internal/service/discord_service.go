package service

import (
	"context"

	"backend/internal/discord"
	"backend/internal/models"
	"backend/internal/repository"
)

type DiscordService interface {
	HandlePing(ctx context.Context, channelID, userID string) error
	AddWhitelist(ctx context.Context, userID, note string) error
	RemoveWhitelist(ctx context.Context, userID string) error
	IsAllowed(ctx context.Context, userID string) (bool, error)
}

type discordService struct {
	whitelist repository.WhitelistRepository
	dsession  discord.Session
}

func NewDiscordService(w repository.WhitelistRepository, d discord.Session) DiscordService {
	return &discordService{
		whitelist: w,
		dsession:  d,
	}
}

func (s *discordService) HandlePing(ctx context.Context, channelID, userID string) error {
	ok, err := s.IsAllowed(ctx, userID)
	if err != nil {
		return err
	}

	msg := "pong"
	if !ok {
		msg = "許可されていないユーザだ"
	}
	_, err = s.dsession.Native().ChannelMessageSend(channelID, msg)
	return err
}

func (s *discordService) AddWhitelist(ctx context.Context, userID, note string) error {
	return s.whitelist.Add(ctx, models.DiscordWhitelist{
		Platform: "discord", UserID: userID, Note: note,
	})
}

func (s *discordService) RemoveWhitelist(ctx context.Context, userID string) error {
	return s.whitelist.Remove(ctx, "discord", userID)
}

func (s *discordService) IsAllowed(ctx context.Context, userID string) (bool, error) {
	return s.whitelist.Exists(ctx, "discord", userID)
}
