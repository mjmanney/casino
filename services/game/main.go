package main

import (
	"casino/libs/store"
	"fmt"
)

func main() {
	st := store.NewEventStore()
	g := NewGame(st)
	fmt.Println("game up", g.State)

	g.StartRound()
	g.PrintHands()

	g.Hit()
	g.PrintHands()

	g.Stand()
	g.DealerPlay()
	g.Settle()

	fmt.Println("round complete", g.State)
	fmt.Println("event log:", st.All())
}
