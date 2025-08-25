package main

import (
	"casino/libs/fsm"
	"casino/libs/store"
)

//	----- Game State Machine -----

/*
Keep the game as the single point of truth.
The game knows the shoe, current state, and active players.
Players should not manipulate the deck or state directly; they should “request” actions that the game validates and executes.
This prevents illegal moves and keeps the event log consistent.
*/
type Game struct {
	State               fsm.State
	Seat1, Seat2, Seat3 *Player
	Dealer              *Dealer
	TurnQueue           []Turn
	Store               *store.EventStore
}

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

// Construct a new Game
func NewGame(store *store.EventStore) *Game {
	d := NewDealer("Dealer")
	return &Game{
		State:     StateTableOpen,
		Seat1:     nil,
		Seat2:     nil,
		Seat3:     nil,
		Dealer:    d,
		TurnQueue: []Turn{},
		Store:     store,
	}
}

// Returns the seats in dealing order, including nils (so we can skip them).
func (g *Game) GetSeats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
}

func (g *Game) DoForEachPlayer(fn func(*Player)) {
	for _, p := range g.GetSeats() {
		if p != nil {
			fn(p)
		}
	}
}

// First build only (table just opened)
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards

	g.Store.Append(store.Event{
		Type:    "Shuffling Cards",
		Payload: map[string]any{"state": g.State},
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

// Call this when the cut card is reached (e.g., between rounds)
func (g *Game) ReshuffleShoe() {
	if g.Dealer.Shoe == nil {
		g.Shuffle()
		return
	}

	g.State = StateShuffleCards
	g.Store.Append(store.Event{
		Type:    "Reshuffling Shoe",
		Payload: map[string]any{"state": g.State},
	})

	// Reshuffle existing shoe contents
	g.Dealer.Shoe.Shuffle(0.65)
	g.State = StateBetsOpen
}

// StartRound initializes a new round
func (g *Game) StartRound() {
	if g.State != StateBetsOpen {
		return
	}
	betAmount := g.Seat1.TotalBet // TODO: Update
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
		hand := NewHand(p.TotalBet, SplitConfig{})
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

	g.EnqueueRoundStart()
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
	for _, p := range g.GetSeats() {
		if p == nil {
			continue
		}
		for _, h := range p.Hands {
			if h.Status != Busted {
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
	player := g.Seat1.Hands[0].Value() // TODO: Update
	dealer := g.Dealer.Hand.Value()
	result := "PUSH"
	if player > 21 || (dealer <= 21 && dealer > player) {
		result = "DEALER WINS"
	} else if dealer > 21 || player > dealer {
		result = "PLAYER WINS"
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
