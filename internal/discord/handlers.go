package discord

import (
	"context"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type PingHandlerFunc func(ctx context.Context, channelID, userID string) error

func RegisterMessageCreate(s Session, handlePing PingHandlerFunc) {
	s.Native().AddHandler(func(ds *discordgo.Session, m *discordgo.MessageCreate) {
		// デバッグログ
		log.Printf(
			"[discord] messageCreate: author=%s bot=%v channel=%s content=%q",
			m.Author.ID,
			m.Author.Bot,
			m.ChannelID,
			m.Content,
		)

		// Bot自身 or 他のBotは無視
		if m.Author == nil || m.Author.Bot {
			return
		}

		// "!ping" で始まるメッセージのみ処理
		if strings.HasPrefix(m.Content, "!ping") {
			log.Printf("[discord] !ping detected from %s in %s", m.Author.ID, m.ChannelID)
			if err := handlePing(context.Background(), m.ChannelID, m.Author.ID); err != nil {
				log.Printf("[discord] HandlePing error: %v", err)
			} else {
				log.Printf("[discord] HandlePing success")
			}
		}
	})
}
