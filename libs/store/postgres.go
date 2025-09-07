package store

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
)

// Database handler that can manage connections
type PostgresStore struct {
	db *sql.DB
}

// Append inserts an event into event_log and returns the assigned id and seq.
// If IdempotencyKey is provided and a duplicate is detected, it returns the
// existing row's id and seq instead of inserting a new row.
func (p *PostgresStore) Append(e Envelope) (int64, int64, error) {
	if p == nil || p.db == nil {
		return 0, 0, errors.New("postgres store is not initialized")
	}

	// Compute next seq for the stream in a transaction to avoid gaps and ensure ordering per stream.
	tx, err := p.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		// Rollback if still active
		_ = tx.Rollback()
	}()

	var existingID, existingSeq sql.NullInt64
	if e.IdempotencyKey != "" {
		// Check for existing event with the same idempotency key
		q := `SELECT id, seq FROM event_log WHERE idempotency_key = $1`
		if err := tx.QueryRow(q, e.IdempotencyKey).Scan(&existingID, &existingSeq); err != nil && err != sql.ErrNoRows {
			return 0, 0, err
		}
		if existingID.Valid {
			// Return existing without inserting
			if err := tx.Commit(); err != nil {
				return 0, 0, err
			}
			return existingID.Int64, existingSeq.Int64, nil
		}
	}

	// Next per-stream sequence
	var nextSeq int64
	if err := tx.QueryRow(`SELECT COALESCE(MAX(seq), 0) + 1 FROM event_log WHERE stream_id = $1`, e.StreamID).Scan(&nextSeq); err != nil {
		return 0, 0, err
	}

	// Insert the row
    insert := `
        INSERT INTO event_log (
            stream_id, stream_type, seq, event_type, payload, schema_version,
            metadata, causation_id, correlation_id, producer, idempotency_key
        ) VALUES ($1,$2,$3,$4, $5::jsonb, $6, $7::jsonb, $8, $9, $10, $11)
        RETURNING id, seq
    `

    // Encode payload and metadata to JSON to avoid driver/map type issues.
    var payloadJSON, metadataJSON []byte
    if e.Payload != nil {
        b, err := json.Marshal(e.Payload)
        if err != nil {
            return 0, 0, fmt.Errorf("marshal payload: %w", err)
        }
        payloadJSON = b
    } else {
        payloadJSON = []byte("{}")
    }
    if e.Metadata != nil {
        b, err := json.Marshal(e.Metadata)
        if err != nil {
            return 0, 0, fmt.Errorf("marshal metadata: %w", err)
        }
        metadataJSON = b
    } else {
        metadataJSON = []byte("{}")
    }
	var id int64
    if err := tx.QueryRow(insert,
        e.StreamID, e.StreamType, nextSeq, e.EventType, payloadJSON, e.SchemaVersion,
        metadataJSON, nullable(e.CausationID), nullable(e.CorrelationID), e.Producer, nullable(e.IdempotencyKey),
    ).Scan(&id, &nextSeq); err != nil {
        return 0, 0, err
    }

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return id, nextSeq, nil
}

// LoadByStream loads events in order for a given stream after a specific seq (exclusive).
func (p *PostgresStore) LoadByStream(streamID string, afterSeq int64, limit int) ([]Envelope, error) {
	if p == nil || p.db == nil {
		return nil, errors.New("postgres store is not initialized")
	}

	base := `SELECT id, stream_id, stream_type, seq, event_type, payload, schema_version, metadata, causation_id, correlation_id, producer, created_at, idempotency_key
             FROM event_log
             WHERE stream_id = $1 AND seq > $2
             ORDER BY seq ASC`
	var rows *sql.Rows
	var err error
	if limit > 0 {
		q := fmt.Sprintf("%s LIMIT %d", base, limit)
		rows, err = p.db.Query(q, streamID, afterSeq)
	} else {
		rows, err = p.db.Query(base, streamID, afterSeq)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Envelope, 0, 64)
	for rows.Next() {
		var e Envelope
		var payload, metadata any
		var causationID, correlationID, idem sql.NullString
		if err := rows.Scan(
			&e.ID, &e.StreamID, &e.StreamType, &e.Seq, &e.EventType, &payload, &e.SchemaVersion, &metadata,
			&causationID, &correlationID, &e.Producer, &e.CreatedAt, &idem,
		); err != nil {
			return nil, err
		}
		// Best-effort type assertions for payload/metadata as map[string]any
		if m, ok := payload.(map[string]any); ok {
			e.Payload = m
		}
		if m, ok := metadata.(map[string]any); ok {
			e.Metadata = m
		}
		e.CausationID = causationID.String
		e.CorrelationID = correlationID.String
		e.IdempotencyKey = idem.String
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
