# Casino

## Project Overview

Casino is an event‑sourced playground for building table games in Go. Each service
records domain events in an append‑only log so that all state—such as player
wallets, table snapshots, and game statistics—is reconstructed from the event
stream. This approach enables reproducibility, temporal queries, and flexible
projections over historical game activity.

### Event‑Sourced Architecture

The core of the system is a Postgres `event_log` table where every domain event
is persisted with metadata like stream identifiers, sequence numbers, and
payload schemas. Services write events to this log and derive read models or
materialized views from it, keeping the log as the single source of truth.

## Prerequisites

- **Go 1.22** – development language and toolchain
- **Docker & Docker Compose** – run Postgres, Redis, Kafka, and Consul locally
- **Flyway CLI** – manage database migrations
- **golangci-lint** *(optional)* – linting for Go code

## Build and Run

The project ships with a Makefile to simplify common tasks:

- `make docker-dev` – start supporting infrastructure via Docker Compose
- `make go-build-services` – compile API and Game services into `bin/`
- `make go-run-services` – run services with `go run`
- `make go-run-all` – run compiled binaries
- `make go-test` – execute unit tests
- `make flyway-migrate` / `make flyway-clean` – apply or reset database migrations

## Blackjack MVP Plan

The initial blackjack Minimum Viable Product (MVP) will be split into two core
services:

1. **API Service** – exposes HTTP endpoints for clients to create tables, join
   games, place bets, and perform actions like hit or stand.
2. **Game Service** – runs the blackjack engine and state machine. It receives
   commands from the API, validates game rules, and appends resulting events to
   the `event_log`.

### Data Flow

1. A player interacts with the **API Service** (HTTP).
2. The API sends a command to the **Game Service** over RPC or messaging.
3. The Game validates the command, updates its finite‑state machine, and
   persists an event to Postgres.
4. Downstream projections or subscribers consume events (via Kafka) to build
   read models, update caches in Redis, or notify clients.
5. Clients query projections through the API for the latest game state.

This modular, event‑driven design enables replayable game history, scalable
read models, and straightforward addition of new games beyond blackjack.
