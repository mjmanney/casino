package store

import "fmt"

// Event represents a domain event.
type Event struct {
	Type    string
	Payload any
}

// EventStore provides a simple in-memory log of events.
type EventStore struct {
	events []Event
}

// NewEventStore creates an empty event store.
func NewEventStore() *EventStore {
	return &EventStore{events: []Event{}}
}

// Append saves an event to the store and prints it.
func (s *EventStore) Append(e Event) {
	s.events = append(s.events, e)
	fmt.Printf("event logged: %s %v\n", e.Type, e.Payload)
}

// All returns all stored events.
func (s *EventStore) All() []Event {
	return s.events
}
