package discord

import (
	"backend/internal/service"

	"github.com/bwmarrin/discordgo"
)

// Router は Discord の Interaction を各処理に振り分ける役割。
type Router struct {
	WhitelistService service.WhitelistService
	// TournamentService service.TournamentService
	// CypherService     service.CypherService
	// BeatService       service.BeatService
}

// NewRouter で必要な service を DI。
func NewRouter(
	whitelistService service.WhitelistService,
	// tournamentService service.TournamentService,
	// cypherService service.CypherService,
	// beatService service.BeatService,
) *Router {
	return &Router{
		WhitelistService: whitelistService,
		// TournamentService: tournamentService,
		// CypherService:     cypherService,
		// BeatService:       beatService,
	}
}

// HandleInteraction は discordgo のイベントハンドラとして登録される入口。
// main.go 側で: session.AddHandler(router.HandleInteraction)
func (r *Router) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {

	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		cmd := CommandName(data.Name)

		switch cmd {
		case CommandPing:
			r.handlePing(s, i)
		case CommandWhitelist:
			r.handleWhitelistPanel(s, i)
		}

	case discordgo.InteractionMessageComponent:
		r.handleWhitelistComponent(s, i)

	case discordgo.InteractionModalSubmit:
		r.handleWhitelistModalSubmit(s, i)
	}
}
