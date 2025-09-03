package blackjack

import (
	"bufio"
	"casino/libs/store"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

//	----- Place Bets -----

/*
Players make their initial bets before cards are dealt.
Basic checks against the player wallet and table min/max.
*/
func (g *Game) PlaceBets() {
	if g.State != StateBetsOpen {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	g.ForEachPlayer(func(p *Player) {
		fmt.Printf("\n%s, Place a wager: ", p.Name)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("input error while reading wager for %s: %v", p.Name, err)
			} else {
				fmt.Printf("no more input (EOF) while reading wager for %s", p.Name)
			}
			return
		}
		cmd := scanner.Text()
		betAmount, err := strconv.Atoi(cmd)
		if err != nil {
			log.Printf("failed to convert %q to int: %v", cmd, err)
			return
		}
		p.Wager(betAmount)
		e := store.Event{
			Type: string(g.State),
			Payload: map[string]any{
				"Player":  p.Name,
				"Wager":   betAmount,
				"RoundID": g.RoundId,
			},
		}
		g.Store.Append(e)

	})
	g.State = StateBetsClosed
}

//	----- Settle -----

/*
Evaluates scores, outcomes and payouts.
*/
func (g *Game) Settle() {
	if g.State != StateBetsSettle {
		return
	}
	dScore := g.Dealer.Hand.Value()

	// Settle Insurance Bets
	g.ForEachPlayer(func(p *Player) {
		h := p.Hands[0]
		if len(h.SideBets) == 0 {
			return
		}

		insuranceSideBet := latestUnpaidSideBet(h.SideBets, InsuranceBet)

		if g.Dealer.Hand.Status == Blackjack && !insuranceSideBet.Paid {
			payout := int(math.Round(float64(insuranceSideBet.Amount) * float64(g.Config.InsurancePayout)))
			payout += insuranceSideBet.Amount
			insuranceSideBet.MarkPaid()
			p.LocalWallet += payout
			g.Store.Append(store.Event{
				Type: string(g.State),
				Payload: map[string]any{
					"BetType":     InsuranceBet,
					"Result":      Win,
					"PlayerID":    p.ID,
					"RoundID":     g.RoundId,
					"WagerAmount": insuranceSideBet.Amount,
					"LocalWallet": p.LocalWallet,
				},
			})
		} else {
			g.Store.Append(store.Event{
				Type: string(g.State),
				Payload: map[string]any{
					"BetType":     InsuranceBet,
					"Result":      Loss,
					"PlayerID":    p.ID,
					"RoundID":     g.RoundId,
					"WagerAmount": insuranceSideBet.Amount,
					"LocalWallet": p.LocalWallet,
				},
			})
		}
	})

	// Settle Main Bets
	g.ForEachPlayer(func(p *Player) {
		for i := range p.Hands {
			h := p.Hands[i]
			wager := h.Bet
			var payout int
			outcome := EvaluateOutcome(h.Value(), h.Status, dScore, g.Dealer.Hand.Status)
			switch outcome {
			case Win:
				if h.Status == Blackjack {
					payout = int(float64(wager) * g.Config.BlackjackPayout)
				} else {
					payout = wager * g.Config.Payout
				}
				payout += wager
				p.LocalWallet += payout
			case Push:
				p.LocalWallet += wager
			case Loss:
				if h.Status == Surrendered {
					wager = wager / 2
					p.LocalWallet += wager
				}
			}

			e := store.Event{
				Type: string(g.State),
				Payload: map[string]any{
					"BetType":     "Standard",
					"Result":      outcome,
					"PlayerID":    p.ID,
					"RoundID":     g.RoundId,
					"WagerAmount": wager,
					"LocalWallet": p.LocalWallet,
				},
			}
			g.Store.Append(e)
		}
	})

	if g.Dealer.Shoe.reshuffle {
		g.ReshuffleShoe()
	}
	g.State = StateBetsOpen
}

// Prompts players for insurance.
func (g *Game) OfferInsurance() {
	fmt.Println("Insurance open.")

	scanner := bufio.NewScanner(os.Stdin)
	g.ForEachPlayer(func(p *Player) {
		fmt.Printf("\n%s, Insurance? (y/n): ", p.Name)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("input error while reading insurance prompt for %s: %v", p.Name, err)
			} else {
				fmt.Printf("no more input (EOF) while reading insurance prompt for %s", p.Name)
			}
			return
		}
		response := scanner.Text()
		var err error
		endTurn := false
		if response == "y" {
			endTurn, err = ApplyAction(g, p.ID, Insurance{}, p.Hands[0])
		}

		if err != nil {
			fmt.Println("error:", err)
		}

		if endTurn {
			_, _ = g.AdvanceTurn()
		}

	})
	fmt.Println("Insurance closed.")
}

// Evaulates and returns player WIN, LOSS or PUSH.
func EvaluateOutcome(pScore int, pHandStatus HandStatus, dScore int, dHandStatus HandStatus) Outcome {
	// Surrender
	if pHandStatus == Surrendered {
		return Loss
	}

	// Blackjack
	if pHandStatus == Blackjack && dHandStatus == Blackjack {
		return Push
	}
	if pHandStatus == Blackjack {
		return Win
	}
	if dHandStatus == Blackjack {
		return Loss
	}

	// Bust
	if pHandStatus == Busted || pScore > 21 {
		return Loss
	}
	if dHandStatus == Busted || dScore > 21 {
		return Win
	}

	// Qualified Hand
	if pScore > dScore {
		return Win
	}
	if pScore < dScore {
		return Loss
	}

	return Push
}
