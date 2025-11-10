package models

type WhitelistPlatform string

const (
	WhitelistPlatformDiscord WhitelistPlatform = "discord"
	WhitelistPlatformVRC     WhitelistPlatform = "vrc"
)

type Whitelist struct {
	ID       uint64
	Platform WhitelistPlatform
	UserID   string // Discord ID など
	Note     string
	// CreatedAt は必要なら追加
}

type WhitelistItem struct {
	ID          uint64
	WhitelistID uint64
	VRCName     string
	VRCNameNorm string
}
