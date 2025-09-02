package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

// The player may surrender their hand, to recover half their original bet, and end their turn.
// Available only on the first action of a turn.
type Surrender struct{}

func (Surrender) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot surrender while in %s", g.State)
	}

	if !h.IsFirstAction() {
		return false, fmt.Errorf("can only surrender on first action")
	}

	h.Surrender()

	e := store.Event{
		Type: "Surrender",
		Payload: map[string]any{
			"PlayerID":    p.ID,
			"Surrendered": true,
		},
	}

	g.Store.Append(e)
	return true, nil
}
