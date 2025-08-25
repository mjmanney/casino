package main

import (
	"fmt"
	"strconv"
)

const MaxHandsPerPlayer = 4

//	----- Player Structures -----

type Player struct {
	ID         string
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

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:       id,
		Name:     name,
		Hands:    make([]*Hand, 0, MaxHandsPerPlayer),
		TotalBet: 0,
		Status:   Active,
	}
}

func (p *Player) Active() { p.Status = Active }
func (p *Player) Idle()   { p.Status = Idle }
func (p *Player) Reset()  { p.Hands = nil; p.Status = Active }

// Add hand to the players collection.
func (p *Player) AddHand(h *Hand) {
	if len(p.Hands) >= MaxHandsPerPlayer {
		fmt.Println("Cannot add more hands: maximum hands per player reached.")
		return
	}
	p.Hands = append(p.Hands, h)
}

// Check that the Player is elligble for SPLIT
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

//	----- Hand Structures -----

// Represents a collection of cards held by a Player or Dealer.
type Hand struct {
	Index      HandIndex
	Cards      []Card
	Status     HandStatus
	Bet        int
	DoubleDown bool
	SplitFrom  HandIndex
}

func NewHand(bet int, opts SplitConfig) *Hand {
	return &Hand{
		Index:      opts.Index,
		Cards:      opts.Cards,
		Status:     Qualified,
		Bet:        bet,
		DoubleDown: false,
		SplitFrom:  opts.SplitFrom,
	}
}

type SplitConfig struct {
	Index     HandIndex
	Cards     []Card
	SplitFrom HandIndex
}

type HandIndex int

type HandStatus int

const (
	Qualified HandStatus = iota
	Busted
	Blackjack
	Surrendered
	Settled
)

func (h *Hand) Qualified() { h.Status = Qualified }
func (h *Hand) Bust()      { h.Status = Busted }
func (h *Hand) Blackjack() { h.Status = Blackjack }
func (h *Hand) Surrender() { h.Status = Surrendered }
func (h *Hand) Settled()   { h.Status = Settled }

// Returns the total value of a hand.
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
