package blackjack

import (
	// Standard libs
	"fmt"
)

// Action is the interface every game command implements.
type Action interface {
	Execute(g *Game, p *Player, h *Hand) (endTurn bool, err error)
}

func ApplyAction(g *Game, pID string, action Action, h *Hand) (bool, error) {
	var p *Player
	switch pID {
	case g.Seat1.ID:
		p = g.Seat1
	case g.Seat2.ID:
		p = g.Seat2
	case g.Seat3.ID:
		p = g.Seat3
	}
	if p == nil {
		return false, fmt.Errorf("unknown player %s", pID)
	}
	return action.Execute(g, p, h)
}
