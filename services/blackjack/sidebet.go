package main

// Holds player side bets, including insurance
type SideBet struct {
	Type   SideBetType
	Amount int
	Paid   bool
}

type SideBetType string

const (
	InsuranceBet  SideBetType = "insurance"
	DealerBustBet SideBetType = "dealer bust"
	PairBet       SideBetType = "pair"
)
