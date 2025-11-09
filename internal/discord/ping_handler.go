package discord

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
)

// /ping コマンドの処理。
// WhitelistService で許可ユーザかどうかを見て返答を変える。
func (r *Router) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// ユーザID取得（Guild / DM 両対応）
	userID := ""
	username := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
		username = i.Member.User.Username
	} else if i.User != nil {
		userID = i.User.ID
		username = i.User.Username
	}

	// ここで誰が叩いたかは必ずログに残す
	log.Printf("[discord] /ping requested by user_id=%s username=%s", userID, username)

	ctx := context.Background()

	ok, err := r.WhitelistService.IsAllowed(ctx, "discord", userID)
	if err != nil {
		// 内部エラーはサーバ側に詳細を吐く
		log.Printf("[discord] /ping whitelist check error: user_id=%s err=%v", userID, err)

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "内部エラーが発生したため、現在ステータスを確認できない。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	msg := "pong"
	flags := discordgo.MessageFlags(0)

	if !ok {
		msg = "許可されていないユーザである。"
		flags = discordgo.MessageFlagsEphemeral
		log.Printf("[discord] /ping denied: user_id=%s username=%s", userID, username)
	} else {
		log.Printf("[discord] /ping ok: user_id=%s username=%s", userID, username)
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	})
}
