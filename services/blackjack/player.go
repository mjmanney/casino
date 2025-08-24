package main

import (
	"fmt"
	"strconv"
)

/*
----- Player Structures -----
Players have a unique ID, name, and hold a collection of Hands.
Players store data about their turn in the game and make requests to the game state.
*/
type Player struct {
	ID         string // unique identifier or seat label
	Name       string
	Hands      []*Hand
	TotalBet   int
	Status     PlayerStatus
	ActiveHand HandIndex
}

type PlayerStatus int

const (
	Active PlayerStatus = iota
	Idle
)

func (p *Player) Active() { p.Status = Active }
func (p *Player) Idle()   { p.Status = Idle }
func (p *Player) Reset()  { p.Hands = nil; p.Status = Active }

// Player Constructor
func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:       id,
		Name:     name,
		Hands:    []*Hand{{}},
		TotalBet: 0,
		Status:   Active,
	}
}

// Returns hands that do not have a BUST or SURRENDER status
func (p *Player) ActiveHands() []*Hand {
	out := make([]*Hand, 0, len(p.Hands))
	for _, h := range p.Hands {
		if h.Status == Qualified || h.Status == Blackjack {
			out = append(out, h)
		}
	}
	return out
}

// Adds a Hand to Player.Hands, used for SPLIT
func (p *Player) AddHand(h *Hand) {
	if len(p.Hands) >= MaxHandsPerPlayer {
		fmt.Println("Cannot add more hands: maximum hands per player reached.")
		return
	}
	p.Hands = append(p.Hands, h)
}

// Check that the Players Hand meets requirements for SPLIT
func (p *Player) CanSplit() bool {
	// Check for the maximum hands
	if len(p.Hands) < MaxHandsPerPlayer {
		c1 := p.Hands[p.ActiveHand].Cards[0].Rank
		c2 := p.Hands[p.ActiveHand].Cards[1].Rank

		// Check for an exact pair
		if c1 == c2 {
			return true
		}
		/*
			Check for two cards that have a value of ten
			but are not an exact pair
		*/
		isValueTen := map[string]bool{
			"10": true,
			"J":  true,
			"Q":  true,
			"K":  true,
		}

		return isValueTen[c1] && isValueTen[c2]
	}
	return false
}

/*
	----- Hand Structures -----
	Hand represents a collection of cards held by a player or dealer.
	It also keeps track of key hands via status, i.e. as Blackjack or Bust.
*/

const MaxHandsPerPlayer = 4

// Hand keeps a collection of cards, and tracks the active hand through an Index.
// Game actions are captured for DOUBLE and SPLIT, as well as the Hand Status.
type Hand struct {
	Index      HandIndex
	Cards      []Card
	Status     HandStatus
	Bet        int
	DoubleDown bool
	SplitFrom  HandIndex
}

// Hand constructor
func NewHand(index int, bet int) *Hand {
	return &Hand{
		Index:      HandIndex(index),
		Cards:      []Card{},
		Status:     Qualified,
		Bet:        bet,
		DoubleDown: false,
		SplitFrom:  -1,
	}
}

type HandIndex int

type HandStatus int

const (
	Qualified HandStatus = iota
	Bust
	Blackjack
	Surrender
	Settled
)

func (h *Hand) Qualified() { h.Status = Qualified }
func (h *Hand) Bust()      { h.Status = Bust }
func (h *Hand) Blackjack() { h.Status = Blackjack }
func (h *Hand) Surrender() { h.Status = Surrender }
func (h *Hand) Settled()   { h.Status = Settled }

// Calculate the hand value, according to Blackjack rules
func (h Hand) Value() int {
	total := 0
	aces := 0
	for _, c := range h.Cards {
		switch c.Rank {
		case "A":
			total += 11
			aces++
		case "K", "Q", "J":
			total += 10
		default:
			v, _ := strconv.Atoi(c.Rank)
			total += v
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}
