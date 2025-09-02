package main

import "fmt"

//	----- Print to STDOUT -----

const asciiPrint = false

func PrintCard(card Card) {
	suitSymbols := map[string]string{
		"Hearts":   "♥",
		"Diamonds": "♦",
		"Clubs":    "♣",
		"Spades":   "♠",
	}

	if !asciiPrint {
		if !card.Hidden {
			fmt.Printf("%s%s", card.Rank, suitSymbols[card.Suit])
		} else {
			// fmt.Println("\U0001F0A0")
			fmt.Printf("🂠")
		}
		return
	}

	if !card.Hidden {
		fmt.Println("+---------+")
		fmt.Printf("| %-6s  |\n", card.Rank)
		fmt.Println("|         |")
		fmt.Printf("|    %-2s   |\n", suitSymbols[card.Suit])
		fmt.Println("|         |")
		fmt.Printf("|%8s |\n", card.Rank)
		fmt.Println("+---------+")
	} else {
		fmt.Println("+---------+")
		fmt.Printf("| %-6s  |\n", "?")
		fmt.Println("|         |")
		fmt.Printf("|    %-2s   |\n", "?")
		fmt.Println("|         |")
		fmt.Printf("|%8s |\n", "?")
		fmt.Println("+---------+")
	}
}

// Prints the hand value and status to STDOUT.
func PrintHand(hand *Hand) {
	c := hand.Cards
	for _, card := range c {
		PrintCard(card)
	}

	switch hand.Status {
	case Blackjack:
		fmt.Print(" ⇒ ", hand.Value(), " BLACKJACK ✪ ")
	case Busted:
		fmt.Print(" ⇒ ", hand.Value(), " BUSTED ✖")
	default:
		fmt.Print(" ⇒ ", hand.Value())
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

// Wrapper on PrintHand that includes player information.
func PrintPlayerHand(player *Player, hand *Hand) {
	fmt.Printf("%s's Hand: ", player.Name)
	PrintHand(hand)
	fmt.Println("")
}

// Wrapper on PrintHand that includes dealer information.
func PrintDealerHand(game *Game) {
	fmt.Print("Dealer Hand: ")
	PrintHand(game.Dealer.Hand)
	fmt.Println("")
}
