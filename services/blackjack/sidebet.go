package main

// Holds player side bets, including insurance.
// All side bets are stored with the initial hand.
type SideBet struct {
	Type   SideBetType
	Amount int
	Paid   bool
}

type SideBetType string

const (
	InsuranceBet  SideBetType = "INSURANCE"
	DealerBustBet SideBetType = "DEALER_BUST"
	PairBet       SideBetType = "PLAYER_PAIR"
)

func NewSideBet(betType SideBetType, bet int) *SideBet {
	return &SideBet{
		Type:   betType,
		Amount: bet,
		Paid:   false,
	}
}

func (s *SideBet) MarkPaid() {
	s.Paid = true
}

func (s *SideBet) IsPaid() bool {
	return s.Paid
}
