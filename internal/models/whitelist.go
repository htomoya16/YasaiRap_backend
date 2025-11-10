package models

import "time"

type WhitelistUser struct {
	ID             uint64
	DiscordUserID  string
	VRCUserID      string
	VRCDisplayName string
	Note           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
