package blackjack

import "fmt"

type Dealer struct {
	Name string
	Shoe *Shoe
	Hand *Hand
}

func NewDealer(name string) *Dealer {
	return &Dealer{
		Name: name,
		Hand: NewHand(0, SplitConfig{}),
		Shoe: &Shoe{},
	}
}

func (d *Dealer) RevealHoleCard() {
	if len(d.Hand.Cards) > 1 {
		fmt.Print("Revealing hidden card...")
		d.Hand.Cards[1].Hidden = false
	}
}

func (d *Dealer) ClearHand() {
	d.Hand = NewHand(0, SplitConfig{})
}

// Dealer checks for insurance and blackjack.
func (g *Game) dealerPeek() {
	for _, card := range g.Dealer.Hand.Cards {
		if g.State != StateDealCards {
			break
		}
		if !card.Hidden {
			switch card.Rank {
			case "A":
				g.State = StateInsuranceTurn
				g.OfferInsurance()
				g.checkBlackjack()
			case "10", "J", "Q", "K":
				g.checkBlackjack()
			default:
				g.State = StatePlayerTurn
			}
		}
	}
}

// Dealer checks for blackjack.
func (g *Game) checkBlackjack() {
	if g.State != StateDealCards && g.State != StateInsuranceTurn {
		return
	}

	fmt.Println("Dealer peeking...")
	if g.Dealer.Hand.ValueAll() == 21 {
		g.Dealer.RevealHoleCard()
		g.Dealer.Hand.Status = Blackjack
		fmt.Println("Dealer has blackjack.")
		g.State = StateBetsSettle
	} else {
		fmt.Println("Dealer does not have blackjack. Resume play.")
		g.State = StatePlayerTurn
	}
}
