package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

// Player may stand to end the turn with a qualifying hand.
type Stand struct{}

func (Stand) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot stand while in %s", g.State)
	}

	if h.Status == Busted {
		return false, fmt.Errorf("stand not applicable; player busted")
	}

	e := store.Event{
		Type: "Stand",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Hand":     h.Cards,
		},
	}
	g.Store.Append(e)
	return true, nil
}
