package main

import (
	"bufio"
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

//	----- Game State Machine -----

/*
Keep the game as the single point of truth.
The game knows the shoe, current state, and active players.
Players should not manipulate the deck or state directly; they should “request” actions that the game validates and executes.
This prevents illegal moves and keeps the event log consistent.
*/
type Game struct {
	State               fsm.State
	Seat1, Seat2, Seat3 *Player
	Dealer              *Dealer
	TurnQueue           []Turn
	Store               *store.EventStore
	Config              *GameConfig
}

type GameConfig struct {
	MinBuyIn        int
	MaxBuyIn        int
	MinWager        int
	MaxWager        int
	Payout          int
	InsurancePayout float64 // 2.0 = 2:1
	BlackjackPayout float64 // 1.5 = 3:2
}

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

func NewGame(store *store.EventStore) *Game {
	d := NewDealer("Dealer")
	return &Game{
		State:     StateTableOpen,
		Seat1:     nil,
		Seat2:     nil,
		Seat3:     nil,
		Dealer:    d,
		TurnQueue: []Turn{},
		Store:     store,
		Config: &GameConfig{
			MinBuyIn:        100,
			MaxBuyIn:        100000,
			MinWager:        5,
			MaxWager:        10000,
			Payout:          1,
			InsurancePayout: 2.0,
			BlackjackPayout: 1.5,
		},
	}
}

//	----- Shuffle -----

/*
Creates new decks and shuffles them together into a single shoe.
See Shoe struct for customizing number of decks and penetration.
*/
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards

	g.Store.Append(store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Minting new decks and shuffling.",
		},
	})

	d1 := NewDeck()
	d2 := NewDeck()
	d3 := NewDeck()
	d4 := NewDeck()
	d5 := NewDeck()
	d6 := NewDeck()

	g.Dealer.Shoe = NewShoe(0.65, d1, d2, d3, d4, d5, d6)

	g.State = StateBetsOpen
}

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
	g.DoForEachPlayer(func(p *Player) {
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
				"Player": p.Name,
				"Wager":  betAmount,
			},
		}
		g.Store.Append(e)

	})
	g.State = StateBetsClosed
}

//	----- Deal Cards -----

/*
Cards are dealt in two passes starting with the players.
After dealing, the dealer and players check for Blackjack.
*/
func (g *Game) DealCards() {
	if g.State != StateBetsClosed {
		return
	}
	g.State = StateDealCards
	e := store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Dealing cards...",
		},
	}
	g.Store.Append(e)

	g.DoForEachPlayer(func(p *Player) {
		h := NewHand(p.TotalBet, SplitConfig{})
		p.AddHand(h)
	})

	for pass := range 2 {
		g.DoForEachPlayer(func(p *Player) {
			card := g.Dealer.Shoe.Draw()
			p.Hands[0].Cards = append(p.Hands[0].Cards, card)
		})
		if pass == 0 {
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, g.Dealer.Shoe.Draw())
		} else {
			card := g.Dealer.Shoe.Draw()
			card.Hidden = true
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
		}
	}
	g.DoForEachPlayer(func(p *Player) {
		if p.Hands[0].checkBlackjack() {
			PrintPlayerHand(p, p.Hands[0])
		}
	})
	g.dealerPeek()
}

//	----- Dealer Turn -----

/*
Dealer's turn begins after all player turns are exhausted from the
games turn queue.
*/
func (g *Game) DealerTurn() {
	if g.State != StateDealerTurn {
		return
	}

	e := store.Event{
		Type: "Dealer Play",
		Payload: map[string]any{
			"state": g.State,
		},
	}
	g.Store.Append(e)

	g.Dealer.RevealHoleCard()
	PrintDealerHand(g)
	if !g.AllPlayersBusted() {
		for g.Dealer.Hand.Value() < 17 {
			card := g.Dealer.Shoe.Draw()
			g.Dealer.Hand.Cards = append(g.Dealer.Hand.Cards, card)
			g.Store.Append(store.Event{Type: "DealerHit", Payload: card})
			PrintDealerHand(g)
		}
	}
	g.State = StateBetsSettle
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
	g.DoForEachPlayer(func(p *Player) {
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
					"WagerAmount": insuranceSideBet.Amount,
					"LocalWallet": p.LocalWallet,
				},
			})
		}
	})

	// Settle Main Bets
	g.DoForEachPlayer(func(p *Player) {
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

type Outcome string

const (
	Win  Outcome = "WIN"
	Loss Outcome = "LOSS"
	Push Outcome = "PUSH"
)

//	----- Game Helpers -----

// Executes a callback function for each player at the table.
func (g *Game) DoForEachPlayer(fn func(*Player)) {
	for _, p := range g.GetSeats() {
		if p != nil {
			fn(p)
		}
	}
}

// Returns the seats in dealing order, including nils.
func (g *Game) GetSeats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
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

// Returns the of most recent unpaid side bet of a given type, otherwise returns nil.
func latestUnpaidSideBet(bets []*SideBet, t SideBetType) *SideBet {
	for i := len(bets) - 1; i >= 0; i-- {
		sb := bets[i]
		if sb == nil {
			continue
		}
		if sb.Type == t && !sb.Paid {
			return sb
		}
	}
	return nil
}

// Returns true if all players have busted.
func (g *Game) AllPlayersBusted() bool {
	for _, p := range g.GetSeats() {
		if p == nil {
			continue
		}
		for _, h := range p.Hands {
			if h.Status != Busted {
				return false
			}
		}
	}
	return true
}

// Prompts players for insurance.
func (g *Game) offerInsurance() {
	fmt.Println("Insurance open.")

	scanner := bufio.NewScanner(os.Stdin)
	g.DoForEachPlayer(func(p *Player) {
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
			endTurn, err = g.ApplyAction(p.ID, Insurance{}, p.Hands[0])
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

// Dealer checks for insurance and blackjack.
func (g *Game) dealerPeek() {
	for _, card := range g.Dealer.Hand.Cards {
		if g.State != StateDealCards {
			break
		}
		if !card.Hidden {
			switch card.Rank {
			case "A":
				g.State = StateInsuranceTurn
				g.offerInsurance()
				g.checkBlackjack()
			case "10", "J", "Q", "K":
				g.checkBlackjack()
			default:
				g.State = StatePlayerTurn
			}
		}
	}
}

// Dealer checks for blackjack.
func (g *Game) checkBlackjack() {
	if g.State != StateDealCards && g.State != StateInsuranceTurn {
		return
	}

	fmt.Println("Dealer peeking...")
	if g.Dealer.Hand.ValueAll() == 21 {
		g.Dealer.RevealHoleCard()
		g.Dealer.Hand.Status = Blackjack
		fmt.Println("Dealer has blackjack.")
		g.State = StateBetsSettle
	} else {
		fmt.Println("Dealer does not have blackjack. Resume play.")
		g.State = StatePlayerTurn
	}
}

// Shuffles the existing shoe.
func (g *Game) ReshuffleShoe() {
	if g.State != StateBetsSettle {
		return
	}
	if g.Dealer.Shoe == nil {
		g.Shuffle()
		return
	}
	g.State = StateShuffleCards
	g.Store.Append(store.Event{
		Type: string(g.State),
		Payload: map[string]any{
			"Message": "Cut card removed - reshuffling.",
		},
	})
	g.Dealer.Shoe.Shuffle(0.65)
}
