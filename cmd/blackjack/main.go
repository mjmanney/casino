package main

import (
	// Standard
	"bufio"
	"flag"
	"fmt"
	"os"

	// Internal
	"casino/libs/store"
	"casino/services/blackjack"
)

func main() {
	// CLI flags for go run ./cmd/blackjack
	dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "Postgres connection string")
	streamID := flag.String("stream", "00000000-0000-0000-0000-000000000001", "Event stream UUID for this table")
	flag.Parse()

	// 1. Initialize event store
	db, err := store.OpenPostgres(*dbURL)
	if err != nil {
		fmt.Println("database error:", err)
		os.Exit(1)
	}
    pst := store.NewPersistentStore(db)
    st := store.NewEventStore(pst, store.EnvelopeDefaults{
        StreamID:      *streamID,
        StreamType:    "table",
        Producer:      "game",
        SchemaVersion: 1,
        Metadata:      map[string]any{"app": "blackjack-cli"},
    })

	// 2. Initialize players
	p1 := blackjack.NewPlayer("1", "Fedor")
	p2 := blackjack.NewPlayer("2", "Shane")
	p3 := blackjack.NewPlayer("3", "Jason")

	// 3. Initialize game, players join
	g := blackjack.NewGame(st)
	g.Join(p1)
	g.Join(p2)
	g.Join(p3)

	// 4. Open table and shuffle cards
	g.Shuffle()

	for {
		// 6. Players place bets before cards are dealt
		g.PlaceBets()
		// 7. Deal cards in two passes
		g.DealCards()
		blackjack.PrintDealerHand(g)
		// 8. Add initial player turns to queue
		if g.State == blackjack.StatePlayerTurn {
			g.ForEachPlayer(func(p *blackjack.Player) {
				h := p.Hands[0]
				blackjack.PrintPlayerHand(p, h)
				turn := blackjack.NewTurn(p, h)
				g.Enqueue(*turn)
			})
		}
		// 9. Process turns from the queue
		for g.State == blackjack.StatePlayerTurn {
			turn, ok := g.Peek()
			p, h := turn.Player, turn.Hand

			if !ok {
				g.State = blackjack.StateDealerTurn
				break
			}
			if p == nil {
				fmt.Println("nil Player in turn; skipping")
				g.AdvanceTurn()
				continue
			}

			if h.Status == blackjack.Blackjack {
				fmt.Println("Player has blackjack; skipping")
				g.AdvanceTurn()
				continue
			}

			blackjack.PrintPlayerHand(p, h)
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
				endTurn, err = blackjack.ApplyAction(g, p.ID, blackjack.Hit{}, h)
			case "s":
				endTurn, err = blackjack.ApplyAction(g, p.ID, blackjack.Stand{}, h)
			case "d":
				endTurn, err = blackjack.ApplyAction(g, p.ID, blackjack.Double{}, h)
			case "sp":
				endTurn, err = blackjack.ApplyAction(g, p.ID, blackjack.Split{}, h)
			case "sur":
				endTurn, err = blackjack.ApplyAction(g, p.ID, blackjack.Surrender{}, h)
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
		// 10. Once the turn queue is empty, the dealer plays their hand.
		g.DealerTurn()

		// 11. Calculate scores, evaluate outcomes and payouts.
		g.Settle()

		// 12. Dev - confirm new round
		fmt.Println("\nPlay another round? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() || scanner.Text() == "y" {
			// Prev Step 5. Clear player data and turn queue
			g.StartRound()
		} else {
			break
		}
	}
}
