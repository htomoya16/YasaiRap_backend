package discord

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Session interface {
	Start(ctx context.Context) error
	Close() error
	Native() *discordgo.Session
}

type session struct {
	dg *discordgo.Session
}

func NewSession() (Session, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is empty")
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	// 必要なIntentを付与（DM/メンション/メッセージ読み取りなど要件次第）
	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages

	return &session{dg: dg}, nil
}

func (s *session) Start(ctx context.Context) error {
	// ハンドラ登録はStart前に行う（外からNative()経由でもOK）
	if err := s.dg.Open(); err != nil {
		return err
	}
	// DiscordのGateway安定待ち（軽いウォームアップ）
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}
	return nil
}

func (s *session) Close() error               { return s.dg.Close() }
func (s *session) Native() *discordgo.Session { return s.dg }
