package main

import (
	"bufio"
	"casino/libs/store"
	"fmt"
	"os"
)

func main() {

	// 1. Initialize event store
	st := store.NewEventStore()

	// 2. Initialize players
	p1 := NewPlayer("1", "Fedor")
	p2 := NewPlayer("2", "Shane")
	p3 := NewPlayer("3", "Jason")

	// 3. Initialize game, players join
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

	// 7. Initialize player turns and start main game loop
	if g.State == StatePlayerTurn {
		g.DoForEachPlayer(func(p *Player) {
			h := p.Hands[0]
			PrintPlayerHand(p, h)
			turn := NewTurn(p, h)
			g.Enqueue(*turn)
		})
	}

	for g.State == StatePlayerTurn {
		turn, ok := g.Peek()
		p, h := turn.Player, turn.Hand

		if !ok {
			g.State = StateDealerTurn
			break
		}
		if p == nil {
			fmt.Println("nil Player in turn; skipping")
			g.AdvanceTurn()
			continue
		}

		if h.Status == Blackjack {
			fmt.Println("Player has blackjack; skipping")
			g.AdvanceTurn()
			continue
		}

		PrintPlayerHand(p, h)
		fmt.Printf("\n%s, Enter action (h)it/(s)tand/(d)ouble/(sp)lit/(sur)render/(q)uit: ", p.Name)
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
			endTurn, err = g.ApplyAction(p.ID, Hit{}, h)
		case "s":
			endTurn, err = g.ApplyAction(p.ID, Stand{}, h)
		case "d":
			endTurn, err = g.ApplyAction(p.ID, Double{}, h)
		case "sp":
			endTurn, err = g.ApplyAction(p.ID, Split{}, h)
		case "sur":
			endTurn, err = g.ApplyAction(p.ID, Surrender{}, h)
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
		if endTurn {
			_, _ = g.AdvanceTurn()
		}
	}

	// 8. Dealer turn
	g.DealerTurn()

	// 9. Calculate scores, evaluate outcomes and payouts
	g.Settle()

}
