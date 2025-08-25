package main

import (
	"fmt"
	"math/rand"
	"time"
)

//	----- Shoe Componenents -----

// A playing card.
type Card struct {
	Suit   string
	Rank   string
	Hidden bool
}

// Holds a collection of cards.
type Deck struct {
	cards []Card
}

// Consists of one or more decks.
type Shoe struct {
	decks    uint
	cards    []Card
	pos      int
	cutIndex int
	rng      *rand.Rand
}

// Creates a new standard 52-card deck.
func NewDeck() *Deck {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

	cards := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			cards = append(cards, Card{Suit: s, Rank: r})
		}
	}

	return &Deck{cards: cards}
}

/*
Creates a shoe of one or more shuffled decks.
Penetratation percentage determines how much of the shoe
is dealt before a new shuffle. Range is capped between 0.65-0.85
with ±0.02 jitter to introduce slight randomness.
*/
func NewShoe(penetration float64, decks ...*Deck) *Shoe {
	if penetration < 0.65 {
		penetration = 0.65
	}
	if penetration > 0.85 {
		penetration = 0.85
	}

	combined := make([]Card, 0, len(decks)*52)
	for _, d := range decks {
		combined = append(combined, d.cards...)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(combined), func(i, j int) { combined[i], combined[j] = combined[j], combined[i] })

	s := &Shoe{
		decks: uint(len(decks)),
		cards: combined,
		pos:   0,
		rng:   r,
	}
	s.placeCutCard(penetration)
	return s
}

// Places the cut card into the shoe.
func (s *Shoe) placeCutCard(penetration float64) {
	n := len(s.cards)
	jitter := (s.rng.Float64()*0.04 - 0.02)
	pct := penetration + jitter
	s.cutIndex = int(pct * float64(n))
}

// Draws a card from the shoe.
func (s *Shoe) Draw() Card {
	if s.ReachedCutCard() {
		// TODO: Finish the current round then reshuffle
	}
	c := s.cards[s.pos]
	s.pos++
	return c
}

// The cut card has been drawn, indicating the final round and a reshuffle.
func (s *Shoe) ReachedCutCard() bool {
	return s.pos >= s.cutIndex
}

// Reshuffles current shoe contents, resets pos and cut card.
func (s *Shoe) Shuffle(penetration float64) {
	s.pos = 0
	s.rng.Shuffle(len(s.cards), func(i, j int) { s.cards[i], s.cards[j] = s.cards[j], s.cards[i] })
	s.placeCutCard(penetration)
}

//	----- Debugging -----

func PrintCard(card Card) {
	suitSymbols := map[string]string{
		"Hearts":   "♥",
		"Diamonds": "♦",
		"Clubs":    "♣",
		"Spades":   "♠",
	}

	fmt.Println("+---------+")
	fmt.Printf("| %-6s  |\n", card.Rank)
	fmt.Println("|         |")
	fmt.Printf("|    %-2s   |\n", suitSymbols[card.Suit])
	fmt.Println("|         |")
	fmt.Printf("|%8s |\n", card.Rank)
	fmt.Println("+---------+")
}

func PrintHand(hand Hand) {
	c := hand.Cards
	for _, card := range c {
		PrintCard(card)
	}
}

func PrintDeck(deck *Deck) {
	for i, card := range deck.cards {
		PrintCard(card)
		if (i+1)%4 == 0 { // Add a line break every 4 cards for better readability
			fmt.Println()
		}
	}
}
