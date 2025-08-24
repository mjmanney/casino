package main

import (
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
)

// Dealer represents the house.
type Dealer struct {
	Shoe *Shoe
	Hand Hand
}

// ----- Game State Machine -----

const (
	StateTableOpen     fsm.State = "TableOpen"
	StateTableClosed   fsm.State = "TableClosed"
	StateShuffleCards  fsm.State = "ShuffleCards"
	StateDealCards     fsm.State = "DealCards"
	StateBetsOpen      fsm.State = "BetsOpen"
	StateBetsClosed    fsm.State = "BetsClosed"
	StateBetsSettle    fsm.State = "BetsSettle"
	StateInsuranceTurn fsm.State = "InsuranceTurn"
	StatePlayerTurn    fsm.State = "PlayerTurn"
	StateDealerTurn    fsm.State = "DealerTurn"
)

/*
Keep the game as the single point of truth.
The game knows the shoe, current state, and active player.
Players should not manipulate the deck or state directly; they should “request” actions that the game validates and executes.
This prevents illegal moves and keeps the event log consistent.
*/
type Game struct {
	State  fsm.State
	Seat1  *Player
	Seat2  *Player
	Seat3  *Player
	Dealer *Dealer
	Store  *store.EventStore
}

// NewGame constructs a Game.
func NewGame(store *store.EventStore) *Game {
	return &Game{
		State:  StateTableOpen,
		Seat1:  nil,
		Seat2:  nil,
		Seat3:  nil,
		Dealer: nil,
		Store:  store,
	}
}

// Returns the seats in dealing order, including nils (so we can skip them).
func (g *Game) seats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
}

func (g *Game) DoForEachPlayer(fn func(*Player)) {
	for _, p := range g.seats() {
		if p != nil {
			fn(p)
		}
	}
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

// Shuffles cards
// This is done when a table state is first opened
// or when the cut card has been dealt
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards

	e := store.Event{
		Type: "Shulffing Cards",
		Payload: map[string]any{
			"state": g.State,
		},
	}
	g.Store.Append(e)
	// Create two new decks
	deck1 := mintDeck()
	deck2 := mintDeck()
	shoe := shuffleDecks(deck1, deck2)
	g.Dealer.Shoe = shoe
	// TODO
	// Logic for deck size and penetration
	// Print deck
	g.State = StateBetsOpen
}

// StartRound initializes a new round
func (g *Game) StartRound() {
	if g.State != StateBetsOpen {
		return
	}
	betAmount := g.Seat1.TotalBet
	e := store.Event{
		Type: "StartRound",
		Payload: map[string]any{
			"state":     g.State,
			"betAmount": betAmount,
		},
	}
	g.Store.Append(e)
	// TODO Player checks for minimum buy in (wallet amount)
	// TODO Player places a wager between table min and max
	// StateBetsOpen lasts for a maximum of 15 seconds
	// Or until player confirms they are ready to play
	g.State = StateBetsClosed
}

// Deal cards to all players at the table.  One card is dealt to each player in order, then the dealer.
// On the second pass, the dealer's hole card is hidden.
func (g *Game) DealCards() {
	if g.State != StateBetsClosed {
		return
	}
	g.State = StateDealCards
	e := store.Event{
		Type: "Dealing Cards",
		Payload: map[string]any{
			"state": g.State,
		},
	}
	g.Store.Append(e)

	g.DoForEachPlayer(func(p *Player) {
		hand := NewHand(0, p.TotalBet)
		p.AddHand(hand)
	})

	for pass := 0; pass < 2; pass++ {
		g.DoForEachPlayer(func(p *Player) {
			card := g.Dealer.Shoe.Draw()
			p.Hands[p.ActiveHand].Cards = append(p.Hands[p.ActiveHand].Cards, card)
		})
		if pass == 0 {
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, g.Dealer.Shoe.Draw())
		} else {
			card := g.Dealer.Shoe.Draw()
			card.Hidden = true
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
		}
	}

	// TODO: Dealer peek() if showing 10 or picture card
	// If blackjack
	//   g.Player.Hand.Blackjack()
	//   g.State = StateBetsSettle
	//
	// TODO: Dealer offers insurance if showing an Ace
	// g.State = InsuranceTurn
	g.State = StatePlayerTurn
}

// DealerPlay draws cards for the dealer per blackjack rules.
func (g *Game) DealerPlay() {
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
	// If all players have busted, dealer only shows hidden card
	// they do not continue to draw
	g.Dealer.Hand.Cards[1].Hidden = false
	if g.AllPlayersBusted() {
		// Skip drawing more cards
	} else {
		for g.Dealer.Hand.Value() < 17 {
			card := g.Dealer.Shoe.Draw()
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
			g.Store.Append(store.Event{Type: "DealerHit", Payload: card})
		}
	}
	g.State = StateBetsSettle
}

func (g *Game) AllPlayersBusted() bool {
	for _, p := range g.seats() {
		if p == nil {
			continue
		}
		for _, h := range p.Hands {
			if h.Status != Bust {
				return false
			}
		}
	}
	return true
}

// Settle evaluates hands and logs the result.
func (g *Game) Settle() {
	if g.State != StateBetsSettle {
		return
	}
	player := g.Seat1.Hands[0].Value()
	dealer := g.Dealer.Hand.Value()
	result := "push"
	if player > 21 || (dealer <= 21 && dealer > player) {
		result = "dealer wins"
	} else if dealer > 21 || player > dealer {
		result = "player wins"
	}
	e := store.Event{
		Type: "Settled",
		Payload: map[string]any{
			"state":  g.State,
			"result": result,
		},
	}
	g.Store.Append(e)
	g.State = StateBetsOpen
}
