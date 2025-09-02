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

// ------- Enqueue ------

// Appends a turn to the end of the queue.
func (g *Game) Enqueue(t Turn) {
	g.TurnQueue = append(g.TurnQueue, t)
	e := store.Event{
		Type: "Enqueue",
		Payload: map[string]any{
			"Message":   "Added turn to end of queue",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}

// ------- Next ------

// Removes and returns the first item from the queue.
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
			"Message":   "Consumed turn",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	})
	return t, true
}

// ------- Peek ------

// Returns the first item from the queue.
func (g *Game) Peek() (Turn, bool) {
	if len(g.TurnQueue) == 0 {
		return Turn{}, false
	}
	return g.TurnQueue[0], true
}

// ------- Advance Turn ------

// Wrapper around Next that processes a turn and advances game state.
func (g *Game) AdvanceTurn() (Turn, bool) {
	t, ok := g.Next()
	if !ok {
		g.State = StateDealerTurn
		g.Store.Append(store.Event{
			Type: "Next",
			Payload: map[string]any{
				"Message":   "Queue empty",
				"QueueSize": 0,
				"NextTurn":  "<none>",
				"GameState": g.State,
			},
		})
		return Turn{}, false
	}

	if len(g.TurnQueue) == 0 {
		g.State = StateDealerTurn
	} else {
		g.State = StatePlayerTurn
	}
	return t, true
}

// ------- Inject Next ------

// Injects a turn to be proccessed next by the queue.  Used for splits.
func (g *Game) InjectNext(t Turn) {
	if len(g.TurnQueue) == 0 {
		g.TurnQueue = append(g.TurnQueue, t)
	} else {
		g.TurnQueue = append(g.TurnQueue, Turn{})
		copy(g.TurnQueue[2:], g.TurnQueue[1:])
		g.TurnQueue[1] = t
	}

	e := store.Event{
		Type: "InjectNext",
		Payload: map[string]any{
			"Message":   "Turn injected",
			"Turn":      t,
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}

// ------- Clear ------

// Removes all turns from the queue.
func (g *Game) Clear() {
	g.TurnQueue = g.TurnQueue[:0]
	e := store.Event{
		Type: "Clear",
		Payload: map[string]any{
			"Message":   "Cleared queue",
			"QueueSize": len(g.TurnQueue),
			"GameState": g.State,
		},
	}
	g.Store.Append(e)
}
