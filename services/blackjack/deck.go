package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ----- Card, Deck, and Shoe Structure -----

// Card represents a playing card.
type Card struct {
	Suit string
	Rank string
}

// Deck holds a collection of cards.
type Deck struct {
	cards []Card
}

// Shoe is a combination of one or more shuffled decks.
type Shoe struct {
	decks uint8
	cards []Card
}

// Creates a new standard 52-card deck.
func mintDeck() *Deck {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

	// Build the deck
	cards := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			cards = append(cards, Card{Suit: s, Rank: r})
		}
	}

	return &Deck{cards: cards}
}

// Shuffles one or more decks into a single Shoe.
func shuffleDecks(decks ...*Deck) *Shoe {
	// Combine all the cards from the passed decks into one slice
	var combinedCards []Card
	for _, deck := range decks {
		combinedCards = append(combinedCards, deck.cards...)
	}

	// Shuffle the combined cards
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(combinedCards), func(i, j int) {
		combinedCards[i], combinedCards[j] = combinedCards[j], combinedCards[i]
	})

	// Return a new Shoe containing the shuffled cards and deck count
	totalDecks := uint8(len(decks))
	return &Shoe{decks: totalDecks, cards: combinedCards}
}

// Removes the top card from the Shoe.
func (s *Shoe) Draw() Card {
	c := s.cards[0]
	s.cards = s.cards[1:]
	return c
}

// ----- Printing -----

// Prints a single card, ASCII-art style
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
	c := hand.cards
	for _, card := range c {
		PrintCard(card)
	}
}

// Print a deck
func PrintDeck(deck *Deck) {
	for i, card := range deck.cards {
		PrintCard(card)
		if (i+1)%4 == 0 { // Add a line break every 4 cards for better readability
			fmt.Println()
		}
	}
}
