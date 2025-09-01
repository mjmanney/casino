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
	p1 := NewPlayer("1", "Fedor")
	p2 := NewPlayer("2", "Shane")
	p3 := NewPlayer("3", "Jason")

	// 3. Initialize Game
	g := NewGame(st)
	g.Seat1 = p1
	g.Seat2 = p2
	g.Seat3 = p3

	// 4. Open table and shuffle cards
	g.Shuffle()

	// 5. Bet loop
	g.PlaceBets()

	// 6. Deal cards
	g.DealCards()
	PrintDealerHand(g)

	// Sets up the initial turn queue
	if g.State == StatePlayerTurn {
		g.DoForEachPlayer(func(p *Player) {
			PrintPlayerHand(p)
			turn := NewTurn(p, p.Hands[0])
			g.Enqueue(*turn)
		})
	}

	// 7. Players Take Action
	for g.State == StatePlayerTurn {
		// 1) Peek current turn without removing it
		turn, ok := g.Peek()
		p := turn.Player

		if !ok {
			// no more turns → move on
			g.State = StateDealerTurn
			break
		}
		if p == nil {
			fmt.Println("nil Player in turn; skipping")
			// Safely remove the bad turn and continue
			g.AdvanceTurn()
			continue
		}

		if p.Hands[p.ActiveHand].Status == Blackjack {
			fmt.Println("Player has blackjack; skipping")
			g.AdvanceTurn()
			continue
		}

		PrintPlayerHand(p)
		fmt.Printf("\nPlayer ID:%s, Enter action (h)it/(s)tand/(d)ouble/(sp)lit/(q)uit: ", p.ID)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading input:", err)
			} else {
				fmt.Println("stdin closed (EOF)")
			}
			return
		}
		cmd := scanner.Text()
		_ = cmd
		var err error
		endTurn := false
		switch cmd {
		case "h":
			endTurn, err = g.ApplyAction(p.ID, Hit{})
		case "s":
			endTurn, err = g.ApplyAction(p.ID, Stand{})
		case "d":
			endTurn, err = g.ApplyAction(p.ID, Double{})
		case "sp":
			endTurn, err = g.ApplyAction(p.ID, Split{})
		case "q":
			fmt.Println("Quitting game.")
			return
		default:
			fmt.Println("Unknown command:", cmd)
			continue
		}
		if err != nil {
			fmt.Println("error:", err)
		}
		PrintPlayerHand(p)

		if endTurn {
			_, _ = g.AdvanceTurn()
		}
	}

	g.DealerPlay()
	g.Settle()

}
