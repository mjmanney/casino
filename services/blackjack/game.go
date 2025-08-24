package main

import (
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
	"strconv"
)

// Dealer represents the house.
type Dealer struct {
	Shoe *Shoe
	Hand Hand
}

// ----- Hand Structures -----
// Hand represents a collection of cards held by a player or dealer.
// It also keeps track of key hands via status, i.e. as Blackjack or Bust.

type Hand struct {
	cards  []Card
	Status HandStatus
}

type HandStatus uint8

const (
	Qualified HandStatus = iota
	Bust
	Blackjack
)

func (h *Hand) Qualified() { h.Status = Qualified }
func (h *Hand) Bust()      { h.Status = Bust }
func (h *Hand) Blackjack() { h.Status = Blackjack }

// Calculate the hand value, according to Blackjack rules
func (h Hand) Value() int {
	total := 0
	aces := 0
	for _, c := range h.cards {
		switch c.Rank {
		case "A":
			total += 11
			aces++
		case "K", "Q", "J", "10":
			total += 10
		default:
			v, _ := strconv.Atoi(c.Rank)
			total += v
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
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

// Game holds the overall game state.
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

// Shuffles cards
// This is done when a table state is first opened
// or when the cut card has been dealt
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards
	g.Store.Append(store.Event{Type: "Shuffling Cards"})
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
	g.Store.Append(store.Event{Type: "Placing Bets", Payload: "<AMOUNT>"})
	// TODO Player checks for minimum buy in (wallet amount)
	// TODO Player places a wager between table min and max
	// StateBetsOpen lasts for a maximum of 15 seconds
	// Or until player confirms they are ready to play
	g.State = StateBetsClosed
}

func (g *Game) DealCards() {
	if g.State != StateBetsClosed {
		return
	}
	g.State = StateDealCards
	g.Store.Append(store.Event{Type: "Dealing Cards"})
	g.Seat1.Hand = Hand{}
	g.Dealer.Hand = Hand{}
	// Player gets a card first, then dealer, then player 2nd card, then dealer 2nd card
	g.Seat1.Hand.cards = append(g.Seat1.Hand.cards, g.Dealer.Shoe.Draw())
	g.Dealer.Hand.cards = append(g.Dealer.Hand.cards, g.Dealer.Shoe.Draw())
	g.Seat1.Hand.cards = append(g.Seat1.Hand.cards, g.Dealer.Shoe.Draw())
	g.Dealer.Hand.cards = append(g.Dealer.Hand.cards, g.Dealer.Shoe.Draw())
	// TODO: Dealer peek() if showing 10 or picture card
	// If blackjack
	//   g.Player.Hand.Blackjack()
	//   g.State = StateBetsSettle
	//
	// TODO: Dealer offers insurance if showing an Ace
	// g.State = InsuranceTurn
	g.State = StatePlayerTurn
}

// Hit deals one card to the player.
func (g *Game) Hit() {
	if g.State != StatePlayerTurn {
		return
	}
	card := g.Dealer.Shoe.Draw()
	g.Seat1.Hand.cards = append(g.Seat1.Hand.cards, card)
	g.Store.Append(store.Event{Type: "Hit", Payload: card})
	if g.Seat1.Hand.Value() > 21 {
		g.Store.Append(store.Event{Type: "PlayerBust"})
		g.Seat1.Hand.Bust()
		g.State = StateDealerTurn
	}
}

// Stand ends the player's turn.
func (g *Game) Stand() {
	if g.State != StatePlayerTurn {
		return
	}
	g.Store.Append(store.Event{Type: "Stand"})
	g.Seat1.Hand.Qualified()
	g.State = StateDealerTurn
}

// DealerPlay draws cards for the dealer per blackjack rules.
func (g *Game) DealerPlay() {
	if g.State != StateDealerTurn {
		return
	}

	// If all players have busted, dealer only shows hidden card
	// they do not continue to draw

	g.Store.Append(store.Event{Type: "DealerPlay"})
	for g.Dealer.Hand.Value() < 17 {
		card := g.Dealer.Shoe.Draw()
		g.Dealer.Hand.cards = append(g.Dealer.Hand.cards, card)
		g.Store.Append(store.Event{Type: "DealerHit", Payload: card})
	}
	g.State = StateBetsSettle
}

// Settle evaluates hands and logs the result.
func (g *Game) Settle() {
	if g.State != StateBetsSettle {
		return
	}
	player := g.Seat1.Hand.Value()
	dealer := g.Dealer.Hand.Value()
	result := "push"
	if player > 21 || (dealer <= 21 && dealer > player) {
		result = "dealer wins"
	} else if dealer > 21 || player > dealer {
		result = "player wins"
	}
	g.Store.Append(store.Event{Type: "Result", Payload: result})
	g.State = StateBetsOpen
}

// PrintHands writes current hands to stdout.
func (g *Game) PrintHands() {
	fmt.Printf("Player hand: %+v (value %d)\n", g.Seat1.Hand, g.Seat1.Hand.Value())
	fmt.Printf("Dealer hand: %+v (value %d)\n", g.Dealer.Hand, g.Dealer.Hand.Value())
}
