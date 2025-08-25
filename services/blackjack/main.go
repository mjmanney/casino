package main

import (
	"casino/libs/store"
	"fmt"
)

func main() {
	st := store.NewEventStore()

	p1 := NewPlayer("1", "PlayerOne")

	g := NewGame(st)
	g.Seat1 = p1

	g.Shuffle()
	p1.TotalBet = 100
	g.StartRound()
	g.DealCards()

	fmt.Println("TurnQueue LENGTH == ", len(g.TurnQueue))

	hit_action := Hit{}
	err := hit_action.Execute(g, g.Seat1)
	if err != nil {
		// handle error
	}

	fmt.Println("TurnQueue LENGTH == ", len(g.TurnQueue))

	stand_action := Stand{}
	err = stand_action.Execute(g, g.Seat1)
	if err != nil {
		// handle error
	}
	PrintHand(*p1.Hands[p1.ActiveHand])
	fmt.Println("TurnQueue LENGTH == ", len(g.TurnQueue))

	g.DealerPlay()
	PrintHand(*g.Dealer.Hand)
	g.Settle()

	fmt.Println("INFO: Round Complete")
	// fmt.Println("INFO: event log:", st.All())
}
