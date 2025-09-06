package blackjack

// Holds player side bets, including insurance.
// All side bets are stored with the initial hand.
type SideBet struct {
	Config SideBetConfig
	Amount int
	Paid   bool
}

type SideBetConfig struct {
	Type     SideBetType
	MinWager int
	MaxWager int
}

type SideBetType string

const (
	InsuranceBet  SideBetType = "INSURANCE"
	DealerBustBet SideBetType = "DEALER_BUST"
	PairBet       SideBetType = "PLAYER_PAIR"
)

func NewSideBet(betConfig SideBetConfig, bet int) *SideBet {
	return &SideBet{
		Config: betConfig,
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

func NewSideBetConfig(betType SideBetType, gc *GameConfig) SideBetConfig {
	// Configure side bet limits as a proportion to the table limits, except for insurance
	var minBet int
	var maxBet int
	if betType == InsuranceBet {
		minBet = 1
		maxBet = gc.MaxWager
	} else {
		calc_min := 0.1 * float64(gc.MinWager)
		minBet = int(calc_min)

		calc_max := 0.5 * float64(gc.MaxWager)
		maxBet = int(calc_max)
	}

	return SideBetConfig{
		Type:     betType,
		MinWager: minBet,
		MaxWager: maxBet,
	}
}
