package discord

import "github.com/bwmarrin/discordgo"

// CommandName は Slash Command 名の型
type CommandName string

// コマンド名一覧（ここだけ見ればOK）
const (
	CommandPing      CommandName = "ping"
	CommandWhitelist CommandName = "whitelist"
	// CommandTournament CommandName = "tournament"
	// CommandCypher     CommandName = "cypher"
	// CommandBeat       CommandName = "beat"
)

// CommandDef は 1コマンド分の定義
type CommandDef struct {
	Name        CommandName
	Description string
	Options     []*discordgo.ApplicationCommandOption
}

// Commands は登録対象のコマンド一覧
// → ApplicationCommandCreate 時にも、ハンドラ側の分岐にもこれを使う。
var Commands = []CommandDef{
	{
		Name:        CommandPing,
		Description: "Botの疎通確認を行う。",
	},
	{
		Name:        CommandWhitelist,
		Description: "自分のホワイトリスト状態を確認・編集する。",
	},
	// 将来的な拡張:
	// {
	// 	Name:        CommandTournament,
	// 	Description: "大会関連の操作を行う。",
	// },
}
