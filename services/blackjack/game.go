package blackjack

import (
	"casino/libs/fsm"
	"casino/libs/store"
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
			MinBuyIn:        100,
			MaxBuyIn:        100000,
			MinWager:        5,
			MaxWager:        10000,
			Payout:          1,
			InsurancePayout: 2.0,
			BlackjackPayout: 1.5,
		},
		RoundId: 0,
	}
}
