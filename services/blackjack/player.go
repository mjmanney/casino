package blackjack

import (
	"fmt"
	"strconv"
)

const MaxHandsPerPlayer = 4

//	----- Player Structures -----

type Player struct {
	ID           string
	Name         string
	Hands        []*Hand
	TotalBet     int
	LocalWallet  int // Bankroll for each game session
	GlobalWallet int // Wallet that persists across game sessions
	Status       PlayerStatus
}

type PlayerStatus int

const (
	Active PlayerStatus = iota
	Idle
)

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:           id,
		Name:         name,
		Hands:        make([]*Hand, 0, MaxHandsPerPlayer),
		TotalBet:     0,
		LocalWallet:  0,
		GlobalWallet: 500000,
		Status:       Active,
	}
}

func (p *Player) Active() { p.Status = Active }
func (p *Player) Idle()   { p.Status = Idle }
func (p *Player) ClearHands() {
	for _, h := range p.Hands {
		h.Cards = nil
		h.SideBets = nil
	}
	p.Hands = p.Hands[:0]
	p.TotalBet = 0
}

// Wager checks to ensure the Player has the funds to make a bet
// and updates the players total bet and local wallet.  Ensure
// to update the Hand's bet property elsewhere.
func (p *Player) Wager(bet int, min int, max int) error {
	if bet > p.LocalWallet {
		return fmt.Errorf("not enough funds in local wallet")
	}

	if bet < min || bet > max {
		return fmt.Errorf("bet must be between %d and %d", min, max)
	}

	p.TotalBet += bet
	p.LocalWallet -= bet
	return nil
}

// Add hand to the players collection.
func (p *Player) AddHand(h *Hand) {
	if len(p.Hands) >= MaxHandsPerPlayer {
		fmt.Println("Cannot add more hands: maximum hands per player reached.")
		return
	}
	p.Hands = append(p.Hands, h)
}

// Check that the Player is elligble for SPLIT
func (player *Player) CanSplit(hand *Hand) (bool, error) {
	if len(player.Hands) >= MaxHandsPerPlayer {
		return false, fmt.Errorf("cannot split; player has maximum number of hands")
	}

	if !hand.IsFirstAction() {
		return false, fmt.Errorf("cannot split; player can only split on first action of hand")
	}

	c1 := hand.Cards[0].Rank
	c2 := hand.Cards[1].Rank

	if c1 == c2 {
		return true, nil
	}

	var match bool
	isValueTen := map[string]bool{
		"10": true,
		"J":  true,
		"Q":  true,
		"K":  true,
	}

	match = isValueTen[c1] && isValueTen[c2]

	if !match {
		return match, fmt.Errorf("cannot split; cards are not same value")
	}

	return match, nil
}

//	----- Hand Structures -----

// Represents a collection of cards held by a Player or Dealer.
type Hand struct {
	Index      HandIndex
	Cards      []Card
	Status     HandStatus
	Bet        int
	SideBets   []*SideBet
	DoubleDown bool
	IsSplit    bool
}

func NewHand(bet int, opts SplitConfig) *Hand {
	return &Hand{
		Index:      opts.Index,
		Cards:      opts.Cards,
		Status:     Qualified,
		Bet:        bet,
		DoubleDown: false,
		IsSplit:    opts.IsSplit,
	}
}

type SplitConfig struct {
	Index   HandIndex
	Cards   []Card
	IsSplit bool
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

// Caclulates the hand total.
// If includeHidden is false, hidden cards are ignored.
func (h Hand) valueCore(includeHidden bool) int {
	total, aces := 0, 0

	for _, c := range h.Cards {
		if c.Hidden && !includeHidden {
			continue
		}
		switch c.Rank {
		case "A":
			total += 11
			aces++
		case "K", "Q", "J":
			total += 10
		default:
			if v, err := strconv.Atoi(c.Rank); err == nil {
				total += v
			}
		}
	}

	// Downgrade aces from 11→1 if value is over 21
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// Returns the total of visible cards only.
func (h Hand) Value() int { return h.valueCore(false) }

// Returns the total including hidden cards.
func (h Hand) ValueAll() int { return h.valueCore(true) }

// Check for players blackjack.
func (h Hand) checkBlackjack() bool {
	if h.Value() == 21 && len(h.Cards) == 2 && !h.IsSplit {
		h.Blackjack()
		return true
	}
	return false
}

// Check that the hand is elligble for actions such as Double or Split.
func (h Hand) IsFirstAction() bool {
	if len(h.Cards) != 2 || h.Status != Qualified {
		return false
	}
	return true
}
