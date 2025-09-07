package blackjack

import "casino/libs/fsm"

type Outcome string

const (
	Win  Outcome = "WIN"
	Loss Outcome = "LOSS"
	Push Outcome = "PUSH"
)

const (
	StateTableOpen     fsm.State = "TableOpen"
	StateTableClosed   fsm.State = "TableClosed"
	StateShuffleCards  fsm.State = "ShuffleCards"
	StateNewRound      fsm.State = "NewRound"
	StateDealCards     fsm.State = "DealCards"
	StateBetsOpen      fsm.State = "BetsOpen"
	StateBetsClosed    fsm.State = "BetsClosed"
	StateBetsSettle    fsm.State = "BetsSettle"
	StateInsuranceTurn fsm.State = "InsuranceTurn"
	StatePlayerTurn    fsm.State = "PlayerTurn"
	StateDealerTurn    fsm.State = "DealerTurn"
)
