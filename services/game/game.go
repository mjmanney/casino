package main

import (
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// ----- card and deck structures -----

// Card represents a playing card.
type Card struct {
	Suit string
	Rank string
}

// Deck holds a collection of cards.
type Deck struct {
	cards []Card
}

// newDeck creates and shuffles a standard 52-card deck.
func newDeck() *Deck {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	cards := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			cards = append(cards, Card{Suit: s, Rank: r})
		}
	}
	d := &Deck{cards: cards}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.cards), func(i, j int) { d.cards[i], d.cards[j] = d.cards[j], d.cards[i] })
	return d
}

// Draw removes the top card from the deck.
func (d *Deck) Draw() Card {
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c
}

// ----- player hand structures -----

// Hand represents a collection of cards held by a player or dealer.
type Hand []Card

// Value calculates the blackjack value of the hand.
func (h Hand) Value() int {
	total := 0
	aces := 0
	for _, c := range h {
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

// Player represents the human participant.
type Player struct {
	Name string
	Hand Hand
}

// Dealer represents the house.
type Dealer struct {
	Deck *Deck
	Hand Hand
}

// ----- game state machine -----

const (
	StateCreated    fsm.State = "Created"
	StateDealing    fsm.State = "Dealing"
	StatePlayerTurn fsm.State = "PlayerTurn"
	StateDealerTurn fsm.State = "DealerTurn"
	StateRoundOver  fsm.State = "RoundOver"
)

// Game holds the overall game state.
type Game struct {
	State  fsm.State
	Player *Player
	Dealer *Dealer
	Store  *store.EventStore
}

// NewGame constructs a Game.
func NewGame(store *store.EventStore) *Game {
	return &Game{
		State:  StateCreated,
		Player: &Player{Name: "Player"},
		Dealer: &Dealer{},
		Store:  store,
	}
}

// StartRound initializes a new round and deals initial cards.
func (g *Game) StartRound() {
	if g.State != StateCreated && g.State != StateRoundOver {
		return
	}
	g.State = StateDealing
	g.Store.Append(store.Event{Type: "StartRound"})
	g.Dealer.Deck = newDeck()
	g.Player.Hand = Hand{}
	g.Dealer.Hand = Hand{}
	g.Player.Hand = append(g.Player.Hand, g.Dealer.Deck.Draw(), g.Dealer.Deck.Draw())
	g.Dealer.Hand = append(g.Dealer.Hand, g.Dealer.Deck.Draw(), g.Dealer.Deck.Draw())
	g.State = StatePlayerTurn
}

// Hit deals one card to the player.
func (g *Game) Hit() {
	if g.State != StatePlayerTurn {
		return
	}
	card := g.Dealer.Deck.Draw()
	g.Player.Hand = append(g.Player.Hand, card)
	g.Store.Append(store.Event{Type: "Hit", Payload: card})
	if g.Player.Hand.Value() > 21 {
		g.Store.Append(store.Event{Type: "PlayerBust"})
		g.State = StateDealerTurn
	}
}

// Stand ends the player's turn.
func (g *Game) Stand() {
	if g.State != StatePlayerTurn {
		return
	}
	g.Store.Append(store.Event{Type: "Stand"})
	g.State = StateDealerTurn
}

// DealerPlay draws cards for the dealer per blackjack rules.
func (g *Game) DealerPlay() {
	if g.State != StateDealerTurn {
		return
	}
	g.Store.Append(store.Event{Type: "DealerPlay"})
	for g.Dealer.Hand.Value() < 17 {
		card := g.Dealer.Deck.Draw()
		g.Dealer.Hand = append(g.Dealer.Hand, card)
		g.Store.Append(store.Event{Type: "DealerHit", Payload: card})
	}
	g.State = StateRoundOver
}

// Settle evaluates hands and logs the result.
func (g *Game) Settle() {
	if g.State != StateRoundOver {
		return
	}
	player := g.Player.Hand.Value()
	dealer := g.Dealer.Hand.Value()
	result := "push"
	if player > 21 || (dealer <= 21 && dealer > player) {
		result = "dealer wins"
	} else if dealer > 21 || player > dealer {
		result = "player wins"
	}
	g.Store.Append(store.Event{Type: "Result", Payload: result})
	g.State = StateCreated
}

// PrintHands writes current hands to stdout.
func (g *Game) PrintHands() {
	fmt.Printf("Player hand: %+v (value %d)\n", g.Player.Hand, g.Player.Hand.Value())
	fmt.Printf("Dealer hand: %+v (value %d)\n", g.Dealer.Hand, g.Dealer.Hand.Value())
}
