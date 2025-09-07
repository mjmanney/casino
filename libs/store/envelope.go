package store

import "time"

// Envelope is the persisted event wrapper recorded in the event log.
// It mirrors the schema in db/migrations/V1__event_log_init.sql.
type Envelope struct {
	ID             int64          // event_log.id (global order)
	StreamID       string         // event_log.stream_id (UUID as string)
	StreamType     string         // event_log.stream_type
	Seq            int64          // event_log.seq (per-stream sequence)
	EventType      string         // event_log.event_type
	Payload        map[string]any // event_log.payload (jsonb)
	SchemaVersion  int            // event_log.schema_version
	Metadata       map[string]any // event_log.metadata (jsonb)
	CausationID    string         // event_log.causation_id (UUID, optional)
	CorrelationID  string         // event_log.correlation_id (UUID, optional)
	Producer       string         // event_log.producer (e.g., "game", "api")
	CreatedAt      time.Time      // event_log.created_at
	IdempotencyKey string         // event_log.idempotency_key (optional)
}

// Appender appends events to a durable store.
type Appender interface {
	Append(e Envelope) (id int64, seq int64, err error)
}

// Loader loads events from the durable store.
type Loader interface {
	// LoadByStream returns events for a stream ordered by seq, starting after
	// the given sequence (use 0 to load from the beginning). Limit <= 0 means no limit.
	LoadByStream(streamID string, afterSeq int64, limit int) ([]Envelope, error)
}

// PostgresStore is a convenience interface combining append and load.
type PersistentStore interface {
	Appender
	Loader
}
