package discord

import (
	"backend/internal/service"
	"context"
	"errors"
	"log"
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

// discord IDの取得
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

	// 現在の紐付け取得（1:1想定）
	link, err := r.WhitelistService.GetDiscordVRC(ctx, userID)
	if err != nil {
		link = nil
	}

	allowed := link != nil
	var names []string
	if link != nil {
		names = []string{link.VRCDisplayName}
	}

	embed := buildWhitelistEmbed(userID, allowed, names)
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

// ボタン定義
func whitelistButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: btnWhitelistRegister,
					Label:    "登録 / 更新",
					Style:    discordgo.PrimaryButton,
				},
				&discordgo.Button{
					CustomID: btnWhitelistDelete,
					Label:    "削除",
					Style:    discordgo.DangerButton,
				},
				&discordgo.Button{
					CustomID: btnWhitelistRefresh,
					Label:    "再表示",
					Style:    discordgo.SecondaryButton,
				},
			},
		},
	}
}

// embedの構築
// names は現状 0 or 1 件想定だが、将来拡張も考えて配列のまま。
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

	fields := []*discordgo.MessageEmbedField{
		{
			Name:  "ステータス",
			Value: status,
		},
	}

	discordValue := "なし"
	if allowed {
		discordValue = "<@" + discordID + ">"
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Discord ID",
		Value: discordValue,
	})

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "紐づいている VRChat 名",
		Value: vrcField,
	})

	return &discordgo.MessageEmbed{
		Title:       "ホワイトリスト状態",
		Description: "この Discord アカウントに紐づいている VRChat アカウントの状態。",
		Color:       color,
		Fields:      fields,
	}
}

// ボタン押下
func (r *Router) handleWhitelistComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	userID := extractUserID(i)
	if userID == "" {
		return
	}

	switch data.CustomID {
	case btnWhitelistRegister:
		r.openWhitelistRegisterModal(s, i)
	case btnWhitelistDelete:
		r.handleWhitelistDelete(s, i, userID)
	case btnWhitelistRefresh:
		r.handleWhitelistRefresh(s, i, userID)
	}
}

// 「登録 / 更新」ボタン → VRChat名入力モーダルを開く
func (r *Router) openWhitelistRegisterModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: modalWhitelistRegister,
			Title:    "VRChat名を登録 / 更新",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    modalInputVRCName,
							Label:       "VRChat 上の表示名（displayName）",
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

// モーダル submit: VRChat displayName から Search All Users → whitelist_users 更新
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
	created, err := r.WhitelistService.RegisterDiscordVRC(ctx, userID, vrcName)

	var msg string
	switch {
	case errors.Is(err, service.ErrInvalidArgument):
		msg = "VRChat名が空か不正。もう一度入力してくれ。"
	case errors.Is(err, service.ErrNoExactMatch):
		msg = "その VRChat名に完全一致するユーザーが見つからなかった。"
	case errors.Is(err, service.ErrMultipleExactMatch):
		msg = "同じ VRChat名のユーザーが複数いるため特定できない。"
	case errors.Is(err, service.ErrAlreadyExists):
		msg = "その VRChatアカウントは既に別の Discord ユーザーに登録されている。"
	case err != nil:
		log.Printf("RegisterDiscordVRC internal error: %+v", err)
		msg = "内部エラーで登録に失敗した。時間をおいて試してくれ。"
	case created:
		msg = "ホワイトリストに登録した。"
	default:
		// 既存行の更新パターン
		msg = "ホワイトリストの情報を更新した。"
	}

	link, _ := r.WhitelistService.GetDiscordVRC(ctx, userID)
	allowed := link != nil
	var names []string
	if link != nil {
		names = []string{link.VRCDisplayName}
	}
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

// 「削除」ボタン: この Discord ユーザーのリンクを物理削除
func (r *Router) handleWhitelistDelete(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	ctx := context.Background()
	err := r.WhitelistService.RemoveDiscord(ctx, userID)

	msg := "ホワイトリストから削除した。"
	if err != nil {
		msg = "内部エラーで削除に失敗した。"
	}

	embed := buildWhitelistEmbed(userID, false, nil)

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

// 「再表示」ボタン: 現在の状態を取り直してEmbed更新
func (r *Router) handleWhitelistRefresh(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	ctx := context.Background()
	link, _ := r.WhitelistService.GetDiscordVRC(ctx, userID)
	allowed := link != nil
	var names []string
	if link != nil {
		names = []string{link.VRCDisplayName}
	}
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
