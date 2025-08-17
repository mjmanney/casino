-- Event log = append-only journal for event-sourced model
-- Purpose: immutable append-only journal of domain events. 
-- Source of truth for an event-sourced game. 
-- All state (wallets, stats, snapshots) is derived from this log.
CREATE TABLE event_log (
  id              BIGSERIAL PRIMARY KEY,               -- global order (monotonic per insert)
  stream_id       UUID        NOT NULL,                -- aggregate id (e.g., table_id)
  stream_type     TEXT        NOT NULL,                -- e.g., 'table', 'wallet', etc. (helps routing/partitioning)
  seq             BIGINT      NOT NULL,                -- per-stream sequence starting at 1
  event_type      TEXT        NOT NULL,                -- e.g., 'RoundStarted','BetPlaced','Settled'
  payload         JSONB       NOT NULL,                -- event body (schema versioned)
  schema_version  INT         NOT NULL DEFAULT 1,      -- payload schema version
  metadata        JSONB       NOT NULL DEFAULT '{}'::jsonb, -- tracing, ip, user-agent, etc.
  causation_id    UUID        NULL,                    -- event id that caused this (for replay/debug)
  correlation_id  UUID        NULL,                    -- ties a request/round together
  producer        TEXT        NOT NULL,                -- 'game', 'api', 'scheduler'
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),  -- write timestamp

  -- Idempotency: dedupe same command/request
  idempotency_key TEXT        NULL,

  -- Guarantees
  UNIQUE (stream_id, seq),
  UNIQUE (idempotency_key) DEFERRABLE INITIALLY IMMEDIATE
);

-- Consume a stream forward
CREATE INDEX IF NOT EXISTS idx_event_stream_seq
  ON event_log (stream_id, seq);

-- Global chronological scans (recovery, backfills)
CREATE INDEX IF NOT EXISTS idx_event_created
  ON event_log (created_at, id);

-- Fast filter by type within a stream
CREATE INDEX IF NOT EXISTS idx_event_stream_type
  ON event_log (stream_id, event_type);

-- Optional: JSON key access if you often filter on payload fields
-- CREATE INDEX IF NOT EXISTS idx_event_payload_gin ON event_log USING GIN (payload jsonb_path_ops);
