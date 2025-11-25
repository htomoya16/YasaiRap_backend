package discord

import (
	"backend/internal/service"
	"context"
	"errors"
	"fmt"
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

// 対象Discordユーザー情報取得（ユーザー名＋アイコンURL）
func extractUserInfo(i *discordgo.InteractionCreate) (id, username, avatarURL string) {
	var u *discordgo.User
	if i.Member != nil && i.Member.User != nil {
		u = i.Member.User
	} else if i.User != nil {
		u = i.User
	}
	if u == nil {
		return "", "", ""
	}
	return u.ID, u.Username, u.AvatarURL("128")
}

// /whitelist 実行時: 状態Embed + ボタン
func (r *Router) handleWhitelistPanel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID, username, avatarURL := extractUserInfo(i)
	if discordID == "" {
		return
	}

	ctx := context.Background()

	// 現在の紐付け取得（1:1想定）
	link, err := r.WhitelistService.GetDiscordVRC(ctx, discordID)
	if err != nil {
		link = nil
	}

	allowed := link != nil

	var (
		names        []string
		vrcAvatarURL string
	)
	if link != nil {
		if link.VRCDisplayName != "" {
			names = []string{link.VRCDisplayName}
		}
		vrcAvatarURL = link.VRCAvatarURL
	}

	embed := buildWhitelistEmbed(discordID, username, avatarURL, allowed, names, vrcAvatarURL)
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
// vrcAvatarURL: whitelist_users に保存した currentAvatarImageUrl を渡す
func buildWhitelistEmbed(
	discordID, username, avatarURL string,
	allowed bool,
	names []string,
	vrcAvatarURL string,
) *discordgo.MessageEmbed {
	var (
		title       string
		description string
		color       int
		statusValue string
	)

	if allowed {
		title = "✅ ホワイトリスト登録済み"
		description = "この Discord アカウントは大会用ホワイトリストに登録されている。"
		color = 0x00cc99
		statusValue = "✅ 登録済み"
	} else {
		title = "❌ ホワイトリスト未登録"
		description = "VRChat 名を登録してホワイトリストに参加できる状態にする必要がある。"
		color = 0xff5555
		statusValue = "❌ 未登録"
	}

	vrcField := "なし"
	if len(names) > 0 {
		vrcField = "- " + strings.Join(names, "\n- ")
	}

	discordValue := "なし"
	if discordID != "" {
		discordValue = "<@" + discordID + ">"
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description + "\n`登録 / 更新` ボタンから VRChat 名を登録・更新できる。",
		Color:       color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    username + " さんのホワイトリスト状態",
			IconURL: avatarURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "ステータス",
				Value: statusValue,
			},
			{
				Name:   "Discord",
				Value:  discordValue,
				Inline: true,
			},
			{
				Name:   "紐づいている VRChat 名",
				Value:  vrcField,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "※このパネルは自分にのみ表示される（Ephemeral）。",
		},
	}

	// VRChatアバター画像（DBに保存された currentAvatarImageUrl）を大きめ表示
	if vrcAvatarURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: vrcAvatarURL,
		}
	}

	return embed
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

	discordID, username, avatarURL := extractUserInfo(i)
	if discordID == "" {
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
	created, err := r.WhitelistService.RegisterDiscordVRC(ctx, discordID, vrcName)

	var msg string
	switch {
	case errors.Is(err, service.ErrInvalidArgument):
		msg = "VRChat名が空か不正。もう一度入力してくれ。"
	case errors.Is(err, service.ErrNoExactMatch):
		msg = "その VRChat名のユーザーはいません。"
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
		msg = "ホワイトリストの情報を更新した。"
	}

	link, _ := r.WhitelistService.GetDiscordVRC(ctx, discordID)
	allowed := link != nil

	var (
		names        []string
		vrcAvatarURL string
	)
	if link != nil {
		if link.VRCDisplayName != "" {
			names = []string{link.VRCDisplayName}
		}
		vrcAvatarURL = link.VRCAvatarURL
	}

	embed := buildWhitelistEmbed(discordID, username, avatarURL, allowed, names, vrcAvatarURL)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: whitelistButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})

	// 登録・更新が成功したときは、同じパネルを公開メッセージとして流す
	if err == nil {
		mention := "<@" + discordID + ">"
		publicMsg := ""
		if created {
			publicMsg = fmt.Sprintf("✅ %s が VRChat アカウント「%s」でホワイトリストに登録された。", mention, vrcName)
		} else {
			publicMsg = fmt.Sprintf("♻️ %s のホワイトリスト情報が更新された。（VRChat: 「%s」）", mention, vrcName)
		}

		_, ferr := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: publicMsg,
			Embeds:  []*discordgo.MessageEmbed{embed}, // /whitelist パネルと同じ内容を公開
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeUsers, // ユーザーだけメンション
				},
			},
		})
		if ferr != nil {
			log.Printf("failed to send public whitelist panel: %+v", ferr)
		}
	}
}

// 「削除」ボタン: この Discord ユーザーのリンクを物理削除
func (r *Router) handleWhitelistDelete(s *discordgo.Session, i *discordgo.InteractionCreate, userID string) {
	_, username, avatarURL := extractUserInfo(i)

	ctx := context.Background()
	err := r.WhitelistService.RemoveDiscord(ctx, userID)

	msg := "ホワイトリストから削除した。"
	if err != nil {
		log.Printf("RemoveDiscord internal error: %+v", err)
		msg = "内部エラーで削除に失敗した。"
	}

	embed := buildWhitelistEmbed(userID, username, avatarURL, false, nil, "")

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
	_, username, avatarURL := extractUserInfo(i)

	ctx := context.Background()
	link, _ := r.WhitelistService.GetDiscordVRC(ctx, userID)
	allowed := link != nil

	var (
		names        []string
		vrcAvatarURL string
	)
	if link != nil {
		if link.VRCDisplayName != "" {
			names = []string{link.VRCDisplayName}
		}
		vrcAvatarURL = link.VRCAvatarURL
	}

	embed := buildWhitelistEmbed(userID, username, avatarURL, allowed, names, vrcAvatarURL)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: whitelistButtons(),
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}
