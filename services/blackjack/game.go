package main

import (
	"bufio"
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
	"log"
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

// Construct a new Game
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

func (g *Game) TableOpen() {
	g.State = StateTableOpen
}

// Returns the seats in dealing order, including nils (so we can skip them).
func (g *Game) GetSeats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
}

func (g *Game) DoForEachPlayer(fn func(*Player)) {
	for _, p := range g.GetSeats() {
		if p != nil {
			fn(p)
		}
	}
}

// First build only (table just opened)
func (g *Game) Shuffle() {
	if g.State != StateTableOpen {
		return
	}
	g.State = StateShuffleCards

	g.Store.Append(store.Event{
		Type: "Shuffle",
		Payload: map[string]any{
			"Message":   "Minting new decks and shuffling.",
			"GameState": g.State,
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

func (g *Game) PlaceBets() {
	if g.State != StateBetsOpen {
		return
	}
	// TODO Player checks for minimum buy in (wallet amount)
	// TODO Player places a wager between table min and max
	// StateBetsOpen lasts for a maximum of 15 seconds
	// Or until player confirms they are ready to play
	scanner := bufio.NewScanner(os.Stdin)
	g.DoForEachPlayer(func(p *Player) {
		fmt.Printf("\n%s, Place a wager: ", p.Name)
		// Block until a line is entered
		if !scanner.Scan() {
			// EOF or input error
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
			Type: "PlacedWager",
			Payload: map[string]any{
				"Player":    p.Name,
				"GameState": g.State,
				"Wager":     betAmount,
			},
		}
		g.Store.Append(e)

	})
	g.State = StateBetsClosed
}

// Call this when the cut card is reached (e.g., between rounds)
func (g *Game) ReshuffleShoe() {
	if g.Dealer.Shoe == nil {
		g.Shuffle()
		return
	}

	g.State = StateShuffleCards
	g.Store.Append(store.Event{
		Type:    "Reshuffling Shoe",
		Payload: map[string]any{"state": g.State},
	})

	// Reshuffle existing shoe contents
	g.Dealer.Shoe.Shuffle(0.65)
	g.State = StateBetsOpen
}

// Deal cards to all players at the table.  One card is dealt to each player in order, then the dealer.
// On the second pass, the dealer's hole card is hidden.
func (g *Game) DealCards() {
	if g.State != StateBetsClosed {
		return
	}
	g.State = StateDealCards
	e := store.Event{
		Type: "Dealing Cards",
		Payload: map[string]any{
			"GameState": g.State,
		},
	}
	g.Store.Append(e)

	g.DoForEachPlayer(func(p *Player) {
		hand := NewHand(p.TotalBet, SplitConfig{})
		p.AddHand(hand)
	})

	for pass := range 2 {
		g.DoForEachPlayer(func(p *Player) {
			card := g.Dealer.Shoe.Draw()
			p.Hands[p.ActiveHand].Cards = append(p.Hands[p.ActiveHand].Cards, card)
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
		p.Hands[p.ActiveHand].checkBlackjack()
	})
	g.dealerPeek()
}

// Dealer checks the hidden card for Blackjack
func (g *Game) checkBlackjack() {
	if g.State != StateDealCards && g.State != StateInsuranceTurn {
		return
	}

	fmt.Println("Dealer peeking...")
	if g.Dealer.Hand.Value() == 21 {
		g.Dealer.RevealHoleCard()
		g.Dealer.Hand.Status = Blackjack
		fmt.Println("Dealer has blackjack.")
		g.State = StateBetsSettle
	} else {
		fmt.Println("Dealer does not have blackjack. Resume play.")
		g.State = StatePlayerTurn
	}
}

/*
Last step after cards are dealt.
Dealer checks their hand for Blackjack.
If dealer Ace is the up card, insurance is offered to players.
Dealer's Blackjack ends the round immedatiely.
*/
func (g *Game) dealerPeek() {
	if g.State != StateDealCards {
		return
	}
	for _, card := range g.Dealer.Hand.Cards {
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

// TODO - add insurance bets
func (g *Game) offerInsurance() {
	fmt.Println("Insurance open.")
	fmt.Println("Insurance closed.")
}

/*
Draws cards for the dealer.
Dealer stands on all 17s.
If all players have busted, the dealer reveals the hole card and does not draw.
*/
func (g *Game) DealerPlay() {
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

// Settle evaluates hands and logs the result.
func (g *Game) Settle() {
	if g.State != StateBetsSettle {
		return
	}
	dScore := g.Dealer.Hand.Value()

	g.DoForEachPlayer(func(p *Player) {
		for h := range p.Hands {
			pHand := p.Hands[h]
			wager := pHand.Bet
			var payout int
			outcome := EvaluateOutcome(pHand.Value(), pHand.Status, dScore, g.Dealer.Hand.Status)
			switch outcome {
			case Win:
				if pHand.Status == Blackjack {
					payout = int(float64(wager) * g.Config.BlackjackPayout)
				} else {
					payout = wager * g.Config.Payout
				}
				payout += wager
				p.LocalWallet += payout
			case Push:
				p.LocalWallet += wager
			}

			e := store.Event{
				Type: "Settled",
				Payload: map[string]any{
					"Result":      outcome,
					"PlayerID":    p.ID,
					"WagerAmount": wager,
					"LocalWallet": p.LocalWallet,
					"GameState":   g.State,
				},
			}
			g.Store.Append(e)
		}
	})
	g.State = StateBetsOpen
}

type Outcome string

const (
	Win  Outcome = "WIN"
	Loss Outcome = "LOSS"
	Push Outcome = "PUSH"
)

func EvaluateOutcome(pScore int, pStatus HandStatus, dScore int, dStatus HandStatus) Outcome {
	// Blackjacks
	if pStatus == Blackjack && dStatus == Blackjack {
		return Push
	}
	if pStatus == Blackjack {
		return Win
	}
	if dStatus == Blackjack {
		return Loss
	}

	// Busts
	if pStatus == Busted || pScore > 21 {
		return Loss
	}
	if dStatus == Busted || dScore > 21 {
		return Win
	}

	// Qualified Hands
	if pScore > dScore {
		return Win
	}
	if pScore < dScore {
		return Loss
	}
	return Push
}
