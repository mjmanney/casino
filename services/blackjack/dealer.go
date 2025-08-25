package main

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
