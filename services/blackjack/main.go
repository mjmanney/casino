package main

import (
	"casino/libs/store"
	"fmt"
)

func main() {
	st := store.NewEventStore()

	p1 := NewPlayer("1", "PlayerOne")
	p2 := NewPlayer("2", "PlayerTwo")
	p3 := NewPlayer("3", "PlayerThree")

	dealer := Dealer{
		Hand: Hand{},
		Shoe: nil,
	}

	g := NewGame(st)
	g.Dealer = &dealer
	fmt.Println("INFO: New game", g.State)
	g.Seat1 = p1
	g.Seat2 = p2
	g.Seat3 = p3
	fmt.Println("INFO: Players joined", g.State)
	g.Shuffle()

	fmt.Println("INFO: Bets open", g.State)
	g.Seat1.Bet = 100
	g.StartRound()
	g.DealCards()

	PrintHand(g.Dealer.Hand)
	PrintHand(g.Seat1.Hand)

	g.Hit()
	PrintHand(g.Seat1.Hand)
	g.Stand()
	g.DealerPlay()
	PrintHand(g.Dealer.Hand)
	g.Settle()

	fmt.Println("INFO: round complete", g.State)
	fmt.Println("INFO: event log:", st.All())
}
