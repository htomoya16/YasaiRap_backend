package discord

import (
	"context"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Serviceに依存しないよう、必要最小の関数型だけ受け取る
type PingHandlerFunc func(ctx context.Context, channelID, userID string) error

func RegisterMessageCreate(d Session, handlePing PingHandlerFunc) {
	d.Native().AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		if strings.HasPrefix(m.Content, "!ping") {
			if err := handlePing(context.Background(), m.ChannelID, m.Author.ID); err != nil {
				log.Printf("HandlePing error: %v", err)
			}
		}
	})
}
