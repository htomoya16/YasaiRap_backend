package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// /ping コマンド。ホワイトリスト登録済みならメッセージを変える。
func (r *Router) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := extractUserID(i)

	allowed := false
	if userID != "" && r.WhitelistService != nil {
		// 新仕様: Discord ID に対応するリンクが存在すればホワイトリスト登録済み
		link, err := r.WhitelistService.GetDiscordVRC(context.Background(), userID)
		if err == nil && link != nil {
			allowed = true
		}
	}

	msg := "pong（ホワイトリスト未登録）"
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
