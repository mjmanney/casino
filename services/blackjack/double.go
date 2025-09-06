package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

/*
The player places an additional bet equal to their original stake.
One card is drawn and ends the turn.  Available only on the first action
of a turn.
*/
type Double struct{}

func (Double) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot double while in %s", g.State)
	}

	if !h.IsFirstAction() {
		return false, fmt.Errorf("can only double on first action")
	}

	err := p.Wager(h.Bet, g.Config.MinWager, g.Config.MaxWager)
	if err != nil {
		fmt.Print("error placing double wager")
	}
	h.DoubleDown = true
	h.Bet += h.Bet
	card := g.Dealer.Shoe.Draw()
	h.Cards = append(h.Cards, card)

	e := store.Event{
		Type: "Double",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"TotalBet": p.TotalBet,
			"Card":     card,
		},
	}
	g.Store.Append(e)
	return true, nil
}
