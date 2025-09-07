package store

import (
	"database/sql"
	"fmt"
)

// Event represents a domain event.
type Event struct {
	Type     string
	Envelope Envelope
	Payload  map[string]any
}

// EventStore provides access to both in-memory and persistant store of events.
type EventStore struct {
	events     []Event
	persistent PersistentStore
	defaults   EnvelopeDefaults
}

func NewEventStore(p PersistentStore, d EnvelopeDefaults) *EventStore {
	return &EventStore{
		events:     []Event{},
		persistent: p,
		defaults:   d,
	}
}

// Creates a new persistant store for Postgres
func NewPersistentStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Append saves an event to the store and prints it.
func (s *EventStore) Append(e Event) {
	s.events = append(s.events, e)
	fmt.Printf("event logged: %s %v\n", e.Type, e.Payload)

	env := Envelope{
		StreamID:      s.defaults.StreamID,
		StreamType:    s.defaults.StreamType,
		EventType:     e.Type,
		Payload:       asMap(e.Payload),
		SchemaVersion: s.defaults.SchemaVersion,
		Metadata:      s.defaults.Metadata,
		CorrelationID: s.defaults.CorrelationId,
		Producer:      s.defaults.Producer,
	}
	if _, _, err := s.persistent.Append(env); err != nil {
		fmt.Printf("persistent append failed: %v\n", err)
	}
}

// All returns all stored events.
func (s *EventStore) All() []Event {
	return s.events
}

type EnvelopeDefaults struct {
	StreamID      string
	StreamType    string
	Producer      string
	SchemaVersion int
	Metadata      map[string]any
	CorrelationId string
}

func asMap(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	// Gracefully handle map[string]string by widening values to any
	if ms, ok := v.(map[string]string); ok {
		out := make(map[string]any, len(ms))
		for k, val := range ms {
			out[k] = val
		}
		return out
	}
	// Fallback: wrap as a value field. Note: ensure nested maps use string keys.
	return map[string]any{"value": v}
}

func (s *EventStore) SetStream(streamID, streamType, producer string) {
	s.defaults.StreamID = streamID
	s.defaults.StreamType = streamType
	s.defaults.Producer = producer
}
func (s *EventStore) SetCorrelationID(cID string) {
	s.defaults.CorrelationId = cID
}
