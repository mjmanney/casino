package main

import (
	"casino/libs/store"
)

type Turn struct {
	Player *Player
	Hand   *Hand
}

func NewTurn(p *Player, h *Hand) *Turn {
	return &Turn{
		Player: p,
		Hand:   h,
	}
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
			"Message":   "Added turn to end of queue.",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}

// Next removes and returns the current turn.
// Returns false if the queue is empty.
func (g *Game) Next() (Turn, bool) {
	if len(g.TurnQueue) == 0 {
		return Turn{}, false
	}

	t := g.TurnQueue[0]
	g.TurnQueue[0] = Turn{}
	g.TurnQueue = g.TurnQueue[1:]

	g.Store.Append(store.Event{
		Type: "Next",
		Payload: map[string]any{
			"Message":   "Consumed turn.",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	})
	return t, true
}

// Peek returns the current turn without advancing.
// Returns false if queue is empty
func (g *Game) Peek() (Turn, bool) {
	if len(g.TurnQueue) == 0 {
		return Turn{}, false
	}
	return g.TurnQueue[0], true
}

// AdvanceTurn consumes one turn and updates game state accordingly.
// If the queue becomes empty after consuming, it's the dealer's turn next;
// otherwise it's still a player's turn.
func (g *Game) AdvanceTurn() (Turn, bool) {
	// Pop
	t, ok := g.Next()
	if !ok {
		g.State = StateDealerTurn
		g.Store.Append(store.Event{
			Type: "AdvanceTurn",
			Payload: map[string]any{
				"Message":       "Queue empty (no turn to pop)",
				"QueueSize":     0,
				"NextTurn":      "<none>",
				"NextGameState": g.State,
			},
		})
		return Turn{}, false
	}

	// Decide next state based on what's LEFT after popping
	if len(g.TurnQueue) == 0 {
		g.State = StateDealerTurn
	} else {
		g.State = StatePlayerTurn
	}

	// Safe "peek" for logging
	nextID := any("<none>")
	if len(g.TurnQueue) > 0 && g.TurnQueue[0].Player != nil {
		nextID = g.TurnQueue[0].Player.ID
	}

	g.Store.Append(store.Event{
		Type: "AdvanceTurn",
		Payload: map[string]any{
			"Message":       "Consumed turn. Next in queue:",
			"NextTurn":      nextID,
			"QueueSize":     len(g.TurnQueue),
			"NextGameState": g.State,
		},
	})
	return t, true
}

/*
LIFO
Used when players create a new hand from SPLIT
*/
func (g *Game) InjectNext(t Turn) {
	// If queue is empty, just append.
	if len(g.TurnQueue) == 0 {
		g.TurnQueue = append(g.TurnQueue, t)
	} else {
		// Optional: sanity check same player as head (split invariant)
		// if head, _ := g.Peek(); head.Player != next.Player { ... }
		g.TurnQueue = append(g.TurnQueue, Turn{}) // grow by one
		copy(g.TurnQueue[2:], g.TurnQueue[1:])    // shift tail right
		g.TurnQueue[1] = t                        // place after head
	}

	e := store.Event{
		Type: "InjectNext",
		Payload: map[string]any{
			"Message":   "Injected turn.",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}

// Build the queue at the start of a round: p1 -> p2 -> p3
func (g *Game) Clear() {
	g.TurnQueue = g.TurnQueue[:0]
	e := store.Event{
		Type: "Clear",
		Payload: map[string]any{
			"Message":   "Emptied the TurnQueue.",
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}
