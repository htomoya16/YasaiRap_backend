package models

type DiscordWhitelist struct {
	ID        int64  `db:"id"`
	Platform  string `db:"platform"` // "discord"
	UserID    string `db:"user_id"`  // DiscordのUserID
	Note      string `db:"note"`
	CreatedAt int64  `db:"created_at"`
}
