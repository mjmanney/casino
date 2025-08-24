package main

import (
	"casino/libs/store"
	"fmt"
)

// Action is the interface every game command implements.
type Action interface {
	Execute(g *Game, p *Player) error
}

/*
HIT: the player draws one card.
After the card is drawn the player may STAND or
continue to HIT if the total hand value is <= 21.
*/
type Hit struct{}

// Execute applies the Hit action to the supplied player.
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

	switch {
	case hand.Value() > 21:
		hand.Bust()
		g.State = StateDealerTurn
	case hand.Value() == 21:
		hand.Qualified()
		g.State = StateDealerTurn
	default:
		hand.Qualified()
		g.State = StatePlayerTurn
	}
	return nil
}

/*
STAND: Players may STAND to end their turn with a qualifying hand.
*/
type Stand struct{}

// Execute applies the Hit action to the supplied player.
func (Stand) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
	}

	hand := p.Hands[p.ActiveHand]

	e := store.Event{
		Type: "Stand",
		Payload: map[string]any{
			"PlayerID": p.ID,
			"Hand":     hand.Cards,
		},
	}
	g.Store.Append(e)
	g.State = StateDealerTurn
	return nil
}

/*
DOUBLE: the player draws one card.  After the card is drawn, their turn is over.
Players may only DOUBLE as the first action of a hand.
Players must match their original bet to double.  If they do not have funds, the action is unavailable.
*/
type Double struct{}

// Execute applies the Double action to the supplied player.
func (Double) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
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
	// Double ends the player turn
	g.State = StateDealerTurn
	g.Store.Append(e)
	return nil
}

/*
SPLIT: the player may SPLIT a hand containing two cards of matching values.
Upon a SPLIT, a single card is dealt to each new hand, and the player progresses through each hand.
Players may SPLIT only as the first action of a hand.
Players may SPLIT up to three times, for a total of four hands.
Players may DOUBLE after splitting.
Players must match their original bet to split.  If they do not have funds, the action is unavailable.
*/
type Split struct{}

// Execute applies the Split action to the supplied player.
func (Split) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
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
	// Create the second hand and add it to the players hands

	newSplitHand := Hand{
		Index:      HandIndex(len(p.Hands)),
		Cards:      []Card{splitCard, spDrawCard2},
		Status:     Qualified,
		Bet:        hand.Bet,
		DoubleDown: false,
		SplitFrom:  p.ActiveHand,
	}
	p.AddHand(&newSplitHand)

	e := store.Event{
		Type: "Split",
		Payload: map[string]any{
			"PlayerID": p.ID,
		},
	}

	g.Store.Append(e)
	return nil
}

/*
INSURANCE: players make may an INSURANCE bet worth half of the original bet.
Players are offered INSURANCE by the dealer only when the dealer up card is an Ace.
If the dealer has Blackjack, the player insurance bet pays 2:1 and their original bet is lost.
Otherwise, the INSURANCE bet is lost and play resumes as normal.
Players may only take INSURANCE after all cards are dealt and before the first player action.
Players must match half the original bet for INSURANCE.  If they do not have funds, the action is unavailable.
*/
type Insurance struct{}

// Execute applies the Split action to the supplied player.
func (Insurance) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
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

/*
  SURRENDER: the player may SURRENDER a forfiet the hand to retain half of their original bet.
  After a SURRENDER the player turn ends.
  Players may SURRENDER only as the first action of a hand.


type Surrender struct{}

// Execute applies the Surrender action to the supplied player.
func (Surrender) Execute(g *Game, p *Player) error {
	if g.State != StatePlayerTurn {
		return fmt.Errorf("cannot hit while in %s", g.State)
	}

	e := store.Event{
		Type: "Surrender",
		Payload: map[string]any{
			"PlayerID": p.ID,
		},
	}

	g.Store.Append(e)
    g.State = StateDealerTurn
	return nil
}
*/
