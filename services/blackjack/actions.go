package main

import (
	"casino/libs/store"
	"fmt"
)

// Action is the interface every game command implements.
type Action interface {
	Execute(g *Game, p *Player, h *Hand) (endTurn bool, err error)
}

func (g *Game) ApplyAction(pID string, action Action, h *Hand) (bool, error) {
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
			"Card":     card,
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

//	----- Stand -----

// Player may STAND to end the turn with a qualifying hand.
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

//	----- Double -----

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

	if !h.isFirstAction() {
		return false, fmt.Errorf("can only double on first action")
	}

	p.Wager(h.Bet)
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

//	----- Split -----

/*
The player places an additional bet equal to their original stake.
Split two cards of matching values, with a single card dealt to each new hand.
Available only on the first action of a turn.  Players can split up to three times.
Doubles allowed after splitting.
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
	p.Wager(splitBetAmount)

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

//	----- Insurance -----

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
	p.Wager(insuranceBetAmount)
	insuranceBet := NewSideBet(InsuranceBet, insuranceBetAmount)
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

//	----- Surrender -----

/*
The player may surrender their hand, to recover half their original bet, and end their turn.
Available only on the first action of a turn.
*/
type Surrender struct{}

func (Surrender) Execute(g *Game, p *Player, h *Hand) (bool, error) {
	if g.State != StatePlayerTurn {
		return false, fmt.Errorf("cannot surrender while in %s", g.State)
	}

	e := store.Event{
		Type: "Surrender",
		Payload: map[string]any{
			"PlayerID": p.ID,
		},
	}

	g.Store.Append(e)
	return true, nil
}
