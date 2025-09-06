package blackjack

import (
	// Standard libs
	"fmt"
	// Internal
	"casino/libs/store"
)

/*
The player places an additional bet equal to their original stake.
Split two cards of matching values, with a single card dealt to each new hand.
Available only on the first action of a turn.  Players can split up to three times
for a total of 4 hands. Doubles allowed after splitting.
*/
type Split struct{}

func (Split) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot split while in %s", g.State)
	}

	canSplit, err := p.CanSplit(h)
	if !canSplit {
		return false, err
	}

	splitBetAmount := h.Bet
	c1 := h.Cards[0]
	c2 := h.Cards[1]
	err = p.Wager(splitBetAmount, h.Bet, h.Bet)
	if err != nil {
		return false, err
	}

	// Active hand becomes just the first card, in a new Cards slice.
	h.Cards = []Card{c1}
	h.IsSplit = true
	h.DoubleDown = false
	h.Status = Qualified

	cardForActiveHand := g.Dealer.Shoe.Draw()
	h.Cards = append(h.Cards, cardForActiveHand)

	// New hand starts with the second card and a turn is injected into the turn queue.
	cardForSplitHand := g.Dealer.Shoe.Draw()
	splitHand := NewHand(splitBetAmount, SplitConfig{
		Index:   HandIndex(len(p.Hands)),
		Cards:   []Card{c2, cardForSplitHand},
		IsSplit: true,
	})

	// Append the new hand and inject its turn to be played next.
	p.AddHand(splitHand)
	g.InjectNext(Turn{
		Player: p,
		Hand:   splitHand,
	})

	g.Store.Append(store.Event{
		Type: "Split",
		Payload: map[string]any{
			"PlayerID":   p.ID,
			"ActiveHand": fmt.Sprintf("%v", h.Cards),
			"SplitHand":  fmt.Sprintf("%v", splitHand.Cards),
		},
	})

	return false, nil
}
