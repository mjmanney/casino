package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

//	----- Hit -----

/*
The player draws one card.
After the card is drawn the player may STAND or
continue to HIT if the total hand value is <= 21.
*/
type Hit struct{}

func (Hit) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot hit while in %s", g.State)
	}

	if h.Status == Busted {
		return false, fmt.Errorf("hit not applicable; player busted")
	}

	card := g.Dealer.Shoe.Draw()
	h.Cards = append(h.Cards, card)

	e := store.Event{
		Type: "Hit",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Card":     card.String(),
		},
	}
	g.Store.Append(e)
	if h.Value() > 21 {
		h.Bust()
		return true, nil
	} else if h.Value() == 21 {
		return true, nil
	}
	return false, nil
}
