package blackjack

import (
	"casino/libs/store"
)

func (g *Game) StartRound() {
	g.RoundId += 1
	g.Clear()
	g.Dealer.ClearHand()
	g.ForEachPlayer(func(p *Player) {
		p.ClearHands()
	})
	g.State = StateBetsOpen
}

//	----- Deal Cards -----

/*
Cards are dealt in two passes starting with the players.
After dealing, the dealer and players check for Blackjack.
*/
func (g *Game) DealCards() {
	if g.State != StateBetsClosed {
		return
	}
	g.State = StateDealCards
	e := store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Dealing cards...",
		},
	}
	g.Store.Append(e)

	g.ForEachPlayer(func(p *Player) {
		h := NewHand(p.TotalBet, SplitConfig{})
		p.AddHand(h)
	})

	for pass := range 2 {
		g.ForEachPlayer(func(p *Player) {
			card := g.Dealer.Shoe.Draw()
			p.Hands[0].Cards = append(p.Hands[0].Cards, card)
		})
		if pass == 0 {
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, g.Dealer.Shoe.Draw())
		} else {
			card := g.Dealer.Shoe.Draw()
			card.Hidden = true
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
		}
	}
	g.ForEachPlayer(func(p *Player) {
		if p.Hands[0].checkBlackjack() {
			PrintPlayerHand(p, p.Hands[0])
		}
	})
	g.dealerPeek()
}

//	----- Dealer Turn -----

/*
Dealer's turn begins after all player turns are exhausted from the
games turn queue.
*/
func (g *Game) DealerTurn() {
	if g.State != StateDealerTurn {
		return
	}

	e := store.Event{
		Type: "Dealer Play",
		Payload: map[string]any{
			"state": g.State,
		},
	}
	g.Store.Append(e)

	g.Dealer.RevealHoleCard()
	PrintDealerHand(g)
	if !g.AllPlayersBusted() {
		for g.Dealer.Hand.Value() < 17 {
			card := g.Dealer.Shoe.Draw()
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
			g.Store.Append(store.Event{Type: "DealerHit", Payload: card})
			if g.Dealer.Hand.Value() > 21 {
				g.Dealer.Hand.Bust()
			}
			PrintDealerHand(g)
		}
	}
	g.State = StateBetsSettle
}

// Shuffles the existing shoe.
func (g *Game) ReshuffleShoe() {
	if g.State != StateBetsSettle {
		return
	}
	if g.Dealer.Shoe == nil {
		g.Shuffle()
		return
	}
	g.State = StateShuffleCards
	g.Store.Append(store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Cut card removed - reshuffling.",
		},
	})
	g.Dealer.Shoe.Shuffle(0.65)
}

//	----- Shuffle -----

/*
Creates new decks and shuffles them together into a single shoe.
See Shoe struct for customizing number of decks and penetration.
*/
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards

	g.Store.Append(store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Minting new decks and shuffling.",
		},
	})

	d1 := NewDeck()
	d2 := NewDeck()
	d3 := NewDeck()
	d4 := NewDeck()
	d5 := NewDeck()
	d6 := NewDeck()

	g.Dealer.Shoe = NewShoe(0.65, d1, d2, d3, d4, d5, d6)

	g.State = StateBetsOpen
}
