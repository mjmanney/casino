package main

import (
	"casino/libs/store"
	"fmt"
)

func main() {
	st := store.NewEventStore()

	p1 := NewPlayer("1", "PlayerOne")

	dealer := Dealer{
		Hand: Hand{},
		Shoe: nil,
	}

	g := NewGame(st)
	g.Seat1 = p1
	g.Dealer = &dealer

	g.Shuffle()
	p1.TotalBet = 100
	g.StartRound()
	g.DealCards()

	hit_action := Hit{}
	err := hit_action.Execute(g, g.Seat1)
	if err != nil {
		// handle error
	}

	stand_action := Stand{}
	err = stand_action.Execute(g, g.Seat1)
	if err != nil {
		// handle error
	}

	g.DealerPlay()
	g.Settle()

	fmt.Println("INFO: Round Complete")
	// fmt.Println("INFO: event log:", st.All())
}
