package main

import (
	"casino/libs/fsm"
	"casino/libs/store"
)

type Turn struct {
	Player *Player
	Hand   *Hand
}

/*
------- Turn Queue Processors ------
*/

// Add a turn to the end of TurnQueue
func (g *Game) Enqueue(t Turn) {
	g.TurnQueue = append(g.TurnQueue, t)
	e := store.Event{
		Type: "Enqueue",
		Payload: map[string]any{
			"enqueuedTurn": t,
			"gameState":    g.State,
		},
	}
	g.Store.Append(e)
}

// Remove the current turn from the TurnQueue
func (g *Game) Dequeue() (Turn, fsm.State) {
	t := g.TurnQueue[0]
	g.TurnQueue = g.TurnQueue[1:]
	if len(g.TurnQueue) == 0 {
		return t, StateDealerTurn
	}
	return t, StatePlayerTurn
}

func (g *Game) AdvanceTurn() {
	t, state := g.Dequeue()
	if state == StateDealerTurn {
		g.State = StateDealerTurn
	}

	e := store.Event{
		Type: "AdvanceTurn",
		Payload: map[string]any{
			"NextStateTurn": state,
			"LastTurn":      t,
		},
	}
	g.Store.Append(e)

	// Mutations here affect the real Player/Hand:
	// e.g., t.Hand.AddCard(...), t.Player.SpendChips(...)
}

/*
LIFO
Used when players create a new hand from SPLIT
*/
func (g *Game) InjectNext(next Turn) {
	// If queue is empty, just append.
	if len(g.TurnQueue) == 0 {
		g.TurnQueue = append(g.TurnQueue, next)
		return
	}
	// Insert at index 1, keeping current head at index 0.
	g.TurnQueue = append(g.TurnQueue, Turn{}) // len+1
	copy(g.TurnQueue[2:], g.TurnQueue[1:])    // shift right from index 1
	g.TurnQueue[1] = next                     // place as the next turn
}

// Build the queue at the start of a round: p1 -> p2 -> p3
func (g *Game) EnqueueRoundStart() {
	e := store.Event{
		Type: "EnqueueRoundStart",
		Payload: map[string]any{
			"message":   "clearing TurnQueue",
			"gameState": g.State,
		},
	}
	g.Store.Append(e)
	g.TurnQueue = g.TurnQueue[:0] // reset for the new round
	enqueuePlayer := func(p *Player) {
		if p == nil {
			return
		}
		g.TurnQueue = append(g.TurnQueue, Turn{Player: p, Hand: p.Hands[p.ActiveHand]})
	}
	enqueuePlayer(g.Seat1)
	enqueuePlayer(g.Seat2)
	enqueuePlayer(g.Seat3)
}
