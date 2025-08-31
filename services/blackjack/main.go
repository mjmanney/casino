package main

import (
	"bufio"
	"casino/libs/store"
	"fmt"
	"os"
)

func main() {

	// 1. Initialize Event Store
	st := store.NewEventStore()

	// 2. Initialize Players
	p1 := NewPlayer("1", "PlayerOne")
	//p2 := NewPlayer("2", "PlayerTwo")
	//p3 := NewPlayer("3", "PlayerThree")

	// 3. Initialize Game
	g := NewGame(st)
	g.Seat1 = p1
	//g.Seat2 = p2
	//g.Seat3 = p3

	// 4. Open table and shuffle cards
	g.Shuffle()

	// 5. Bet loop
	p1.TotalBet = 100

	// 6. Round
	g.StartRound()

	// 7. Deal cards
	g.DealCards()

	fmt.Println("Player hand:")
	PrintHand(*p1.Hands[p1.ActiveHand])
	fmt.Println("Dealer hand:")
	PrintHand(*g.Dealer.Hand)

	scanner := bufio.NewScanner(os.Stdin)
	for g.State == StatePlayerTurn {
		fmt.Print("Enter action (h)it/(s)tand/(d)ouble/(sp)lit/(q)uit: ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			} else {
				fmt.Println("stdin closed (EOF)")
			}
			return
		}
		cmd := scanner.Text() // requires trailing Enter
		_ = cmd
		var err error
		switch cmd {
		case "h":
			err = g.ApplyAction(p1.ID, Hit{})
		case "s":
			err = g.ApplyAction(p1.ID, Stand{})
		case "d":
			err = g.ApplyAction(p1.ID, Double{})
		case "sp":
			err = g.ApplyAction(p1.ID, Split{})
		case "q":
			fmt.Println("Quitting game.")
			return
		default:
			fmt.Println("unknown command:", cmd)
			continue
		}
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println("Player hand:")
		PrintHand(*p1.Hands[p1.ActiveHand])
	}

	g.DealerPlay()
	g.Settle()

}
