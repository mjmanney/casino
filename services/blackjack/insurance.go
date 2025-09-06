package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

/*
The player places an additional bet equal to half their original stake.
Available only when the dealer up card is an Ace.
On dealer Blackjack, insurance pays 2:1 and the original stake is lost.
Otherwise, the insurance bet is lost and play resumes as normal.
Offered after all cards are dealt and before the first player action.
*/
type Insurance struct{}

func (Insurance) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StateInsuranceTurn {
		return false, fmt.Errorf("cannot accept insurance while in %s", g.State)
	}

	insuranceBetAmount := h.Bet / 2
	InsuranceSideBetConfig := NewSideBetConfig(InsuranceBet, g.Config)
	insuranceBet := NewSideBet(InsuranceSideBetConfig, insuranceBetAmount)
	err := p.Wager(insuranceBet.Amount, insuranceBet.Config.MinWager, InsuranceSideBetConfig.MaxWager)
	if err != nil {
		return false, fmt.Errorf("error placing insurance bet")
	}
	h.SideBets = append(h.SideBets, insuranceBet)

	e := store.Event{
		Type: "Insurance",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Insured":  true,
			"Amount":   insuranceBetAmount,
		},
	}

	g.Store.Append(e)
	return false, nil
}
