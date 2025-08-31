package main

import (
	"casino/libs/store"
	"fmt"
)

// Action is the interface every game command implements.
type Action interface {
	Execute(g *Game, p *Player) error
}

func (g *Game) ApplyAction(playerID string, a Action) error {
	var p *Player
	switch playerID {
	case g.Seat1.ID:
		p = g.Seat1
	case g.Seat2.ID:
		p = g.Seat2
	case g.Seat3.ID:
		p = g.Seat3
	}
	if p == nil {
		return fmt.Errorf("unknown player %s", playerID)
	}
	return a.Execute(g, p)
}

//	----- Hit -----

/*
The player draws one card.
After the card is drawn the player may STAND or
continue to HIT if the total hand value is <= 21.
*/
type Hit struct{}

func (Hit) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
	}

	hand := p.Hands[p.ActiveHand]
	card := g.Dealer.Shoe.Draw()
	hand.Cards = append(hand.Cards, card)

	e := store.Event{
		Type: "Hit",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Card":     card,
		},
	}
	g.Store.Append(e)
	if hand.Value() > 21 {
		hand.Bust()
		g.AdvanceTurn()
	} else if hand.Value() == 21 {
		g.AdvanceTurn()
	}
	return nil
}

//	----- Stand -----

// Player may STAND to end the turn with a qualifying hand.
type Stand struct{}

func (Stand) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot stand while in %s", g.State)
	}

	hand := p.Hands[p.ActiveHand]

	if hand.Status == Busted {
		return fmt.Errorf("stand not applicable; player busted")
	}

	e := store.Event{
		Type: "Stand",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Hand":     hand.Cards,
		},
	}
	g.Store.Append(e)
	g.AdvanceTurn()
	return nil
}

//	----- Double -----

/*
The player places an additional bet equal to their original stake.
One card is drawn and ends the turn.  Available only on the first action
of a turn.
*/
type Double struct{}

func (Double) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot double while in %s", g.State)
	}

	hand := p.Hands[p.ActiveHand]
	card := g.Dealer.Shoe.Draw()
	hand.Cards = append(hand.Cards, card)

	e := store.Event{
		Type: "Double",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Card":     card,
		},
	}
	g.Store.Append(e)
	g.AdvanceTurn()
	return nil
}

//	----- Split -----

/*
The player places an additional bet equal to their original stake.
Split two cards of matching values, with a single card dealt to each new hand.
Available only on the first action of a turn.  Players can split up to three times.
Doubles allowed after splitting.
*/
type Split struct{}

// Execute applies the Split action to the supplied player.
func (Split) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot split while in %s", g.State)
	}
	if !p.CanSplit() {
		return fmt.Errorf("cannot split due to maximum number of hands or different card values")
	}

	// Get the active hand
	hand := p.Hands[p.ActiveHand]
	// Pop the last card from the current hand
	splitCard := hand.Cards[len(hand.Cards)-1]
	// Draw a card from the shoe
	spDrawCard1 := g.Dealer.Shoe.Draw()
	// Add the card to the active hand
	hand.Cards = append(hand.Cards, spDrawCard1)
	// Draw another card from the shoe
	spDrawCard2 := g.Dealer.Shoe.Draw()
	// Create the new split hand
	splitConfig := SplitConfig{
		Index:     HandIndex(len(p.Hands)),
		Cards:     []Card{splitCard, spDrawCard2},
		SplitFrom: p.ActiveHand,
	}
	newSplitHand := NewHand(hand.Bet, splitConfig)
	p.AddHand(newSplitHand)
	g.InjectNext(Turn{
		Player: p,
		Hand:   newSplitHand,
	})
	e := store.Event{
		Type: "Split",
		Payload: map[string]any{
			"PlayerID": p.ID,
		},
	}
	g.Store.Append(e)
	return nil
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

func (Insurance) Execute(g *Game, p *Player) error {
	if g.State != StateInsuranceTurn {
		return fmt.Errorf("cannot accept insurance while in %s", g.State)
	}

	/*
	   TODO:
	*/

	e := store.Event{
		Type: "Insurance",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Insured":  false,
			"Amount":   0,
		},
	}

	g.Store.Append(e)
	return nil
}

//	----- Surrender -----

/*
The player may surrender their hand, to recover half their original bet, and end their turn.
Available only on the first action of a turn.
*/
type Surrender struct{}

func (Surrender) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot surrender while in %s", g.State)
	}

	e := store.Event{
		Type: "Surrender",
		Payload: map[string]any{
			"PlayerID": p.ID,
		},
	}

	g.Store.Append(e)
	g.AdvanceTurn()
	return nil
}
