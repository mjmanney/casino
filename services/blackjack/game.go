package blackjack

import (
	"casino/libs/fsm"
	"casino/libs/store"
	"fmt"
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
	RoundId             int
}

type GameConfig struct {
	MinBuyIn        int
	MaxBuyIn        int
	MinWager        int
	MaxWager        int
	Payout          int
	InsurancePayout float64 // 2.0 = 2:1
	BlackjackPayout float64 // 1.5 = 3:2
	SurrenderPayout float64 // 0.5 = 1:2
}

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
			MinBuyIn:        1000,
			MaxBuyIn:        100000,
			MinWager:        250,
			MaxWager:        10000,
			Payout:          1,
			InsurancePayout: 2.0,
			BlackjackPayout: 1.5,
			SurrenderPayout: 0.5,
		},
		RoundId: 0,
	}
}

// Player joins the game if there is an open seat and they have enough funds.
func (g *Game) Join(p *Player) error {
	if g.Seat1 != nil && g.Seat2 != nil && g.Seat3 != nil {
		return fmt.Errorf("table is full")
	}

	if p.GlobalWallet < g.Config.MinBuyIn {
		return fmt.Errorf("not enough funds to join the table")
	}

	if p.GlobalWallet >= g.Config.MaxBuyIn {
		funds := p.GlobalWallet - g.Config.MaxBuyIn
		p.GlobalWallet -= funds
		p.LocalWallet = funds
	} else {
		funds := p.GlobalWallet
		p.GlobalWallet -= funds
		p.LocalWallet = funds
	}

	if g.Seat1 == nil {
		g.Seat1 = p
	} else if g.Seat2 == nil {
		g.Seat2 = p
	} else if g.Seat3 == nil {
		g.Seat3 = p
	}
	return nil
}

// Player leaves the game, transferring their local wallet back to their global wallet.
func (g *Game) Leave(p *Player) error {
	if g.Seat1 == p {
		g.Seat1 = nil
	} else if g.Seat2 == p {
		g.Seat2 = nil
	} else if g.Seat3 == p {
		g.Seat3 = nil
	} else {
		return fmt.Errorf("player not at table")
	}

	p.GlobalWallet += p.LocalWallet
	p.LocalWallet = 0

	return nil
}

// Returns the seats in dealing order, including nils.
func (g *Game) Seats() []*Player {
	return []*Player{g.Seat1, g.Seat2, g.Seat3}
}

// Executes a function for each non-nil player at the table.
func (g *Game) ForEachPlayer(fn func(*Player)) {
	for _, p := range g.Seats() {
		if p != nil {
			fn(p)
		}
	}
}
