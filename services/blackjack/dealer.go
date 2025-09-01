package main

import "fmt"

type Dealer struct {
	Name string
	Shoe *Shoe
	Hand *Hand
}

func NewDealer(name string) *Dealer {
	return &Dealer{
		Name: name,
		Hand: &Hand{
			Cards: []Card{},
		},
		Shoe: &Shoe{},
	}
}

func (d *Dealer) RevealHoleCard() {
	if len(d.Hand.Cards) > 1 {
		fmt.Print("Revealing hidden card...")
		d.Hand.Cards[1].Hidden = false
	}
}
