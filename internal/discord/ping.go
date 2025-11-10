package discord

import (
	"backend/internal/models"
	"context"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := extractUserID(i)

	allowed := false
	if userID != "" && r.WhitelistService != nil {
		ok, err := r.WhitelistService.IsAllowed(context.Background(), models.WhitelistPlatformDiscord, userID)
		if err == nil {
			allowed = ok
		}
	}

	msg := "pong"
	if allowed {
		msg = "pong（ホワイトリスト登録済み）"
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
