// internal/discord/whitelist.go (の先頭あたり)
package discord

import (
	"backend/internal/models"
	"backend/internal/service"
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	btnWhitelistRegister = "wl_register"
	btnWhitelistDelete   = "wl_delete"
	btnWhitelistRefresh  = "wl_refresh"

	modalWhitelistRegister = "wl_modal_register"
	modalInputVRCName      = "wl_modal_input_vrc_name"
)

func extractUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

// /whitelist 実行時: 状態Embed + ボタン
func (r *Router) handleWhitelistPanel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := extractUserID(i)
	if userID == "" {
		return
	}

	ctx := context.Background()

	allowed, err := r.WhitelistService.IsAllowed(ctx, models.WhitelistPlatformDiscord, userID)
	if err != nil {
		allowed = false
	}

	names, err := r.WhitelistService.GetDiscordVRCNames(ctx, userID)
	if err != nil {
		names = nil
	}

	// embed
	embed := buildWhitelistEmbed(userID, allowed, names)
	// buttons
	components := whitelistButtons()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// buttonの種類
func whitelistButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: btnWhitelistRegister,
					Label:    "登録 / 追加",
					Style:    discordgo.PrimaryButton,
				},
				&discordgo.Button{
					CustomID: btnWhitelistDelete,
					Label:    "削除",
					Style:    discordgo.DangerButton,
				},
				&discordgo.Button{
					CustomID: btnWhitelistRefresh,
					Label:    "更新",
					Style:    discordgo.SecondaryButton,
				},
			},
		},
	}
}

// embedの構築
func buildWhitelistEmbed(discordID string, allowed bool, names []string) *discordgo.MessageEmbed {
	status := "未登録"
	color := 0xff9933
	if allowed {
		status = "登録済み"
		color = 0x00cc99
	}

	vrcField := "なし"
	if len(names) > 0 {
		vrcField = "- " + strings.Join(names, "\n- ")
	}

	return &discordgo.MessageEmbed{
		Title:       "ホワイトリスト状態",
		Description: "このDiscordアカウントのホワイトリスト登録状況。",
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "ステータス",
				Value: status,
			},
			{
				Name:  "Discord ID",
				Value: discordID,
			},
			{
				Name:  "紐づいているVRChat名",
				Value: vrcField,
			},
		},
	}
}

// ボタン押下
func (r *Router) handleWhitelistComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	userID := extractUserID(i)
	if userID == "" {
		return
	}

	// data.CustomIDが
	switch data.CustomID {
	// "登録"なら
	case btnWhitelistRegister:
		r.openWhitelistRegisterModal(s, i)
	// "削除"なら
	case btnWhitelistDelete:
		r.handleWhitelistDeleteAll(s, i, userID)
	// "更新"なら
	case btnWhitelistRefresh:
		r.handleWhitelistRefresh(s, i, userID)
	}
}

// 登録
func (r *Router) openWhitelistRegisterModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: modalWhitelistRegister,
			Title:    "VRChat名を登録 / 追加",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    modalInputVRCName,
							Label:       "VRChat上の名前",
							Style:       discordgo.TextInputShort,
							Required:    true,
							Placeholder: "例: 野菜ラップ",
						},
					},
				},
			},
		},
	})
}

// モーダルでsubmitしたら
func (r *Router) handleWhitelistModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	if data.CustomID != modalWhitelistRegister {
		return
	}

	userID := extractUserID(i)
	if userID == "" {
		return
	}

	var vrcName string

	// モーダル内の TextInput を正しく拾う
	for _, comp := range data.Components {
		row, ok := comp.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, inner := range row.Components {
			input, ok := inner.(*discordgo.TextInput)
			if !ok {
				continue
			}
			if input.CustomID == modalInputVRCName {
				vrcName = strings.TrimSpace(input.Value)
			}
		}
	}

	ctx := context.Background()
	// userIDとvrcnameでデータベース登録のserviceへ
	created, err := r.WhitelistService.RegisterDiscordVRC(ctx, userID, vrcName)

	var msg string
	switch {
	case err == service.ErrInvalidArgument:
		msg = "VRChat名が空か不正だ。入力を確認してくれ。"
	case err == service.ErrAlreadyExists:
		msg = "そのVRChat名はすでに登録済みだ。"
	case err != nil:
		msg = "内部エラーで登録に失敗した。"
	case created:
		msg = "ホワイトリストに登録した。"
	default:
		msg = "処理は完了した可能性が高いが、状態を確認してくれ。"
	}

	allowed, _ := r.WhitelistService.IsAllowed(ctx, models.WhitelistPlatformDiscord, userID)
	names, _ := r.WhitelistService.GetDiscordVRCNames(ctx, userID)
	embed := buildWhitelistEmbed(userID, allowed, names)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: whitelistButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (r *Router) handleWhitelistDeleteAll(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	ctx := context.Background()
	err := r.WhitelistService.Remove(ctx, models.WhitelistPlatformDiscord, userID)

	msg := "ホワイトリストから削除した。"
	if err != nil {
		msg = "内部エラーで削除できなかった。"
	}

	allowed := false
	names := []string{}
	embed := buildWhitelistEmbed(userID, allowed, names)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: whitelistButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (r *Router) handleWhitelistRefresh(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	ctx := context.Background()
	allowed, _ := r.WhitelistService.IsAllowed(ctx, models.WhitelistPlatformDiscord, userID)
	names, _ := r.WhitelistService.GetDiscordVRCNames(ctx, userID)
	embed := buildWhitelistEmbed(userID, allowed, names)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: whitelistButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}
