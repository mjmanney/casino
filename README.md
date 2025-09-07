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
- `make db-health`– verifies postgres connection using cmd/tools/dbhealth


## Blackjack MVP Plan

The initial blackjack Minimum Viable Product (MVP) will be split into two core
services:

1. **API Service** – exposes HTTP endpoints for clients to create tables, join
   games, place bets, and perform actions like hit or stand.
2. **Game Service** – runs the blackjack engine and state machine. It receives
   commands from the API, validates game rules, and appends resulting events to
  the `event_log`.

### Persistent Event Store (CLI Mirroring)

The in-memory event store used by the Blackjack CLI can optionally mirror events to Postgres. This is useful during development to persist the game log without changing gameplay code.

Run the Blackjack CLI with persistence enabled:

```
go run ./cmd/blackjack -persist -db "$DATABASE_URL" -stream 00000000-0000-0000-0000-000000000001
```

Flags:

- `-persist`: enable Postgres mirroring
- `-db`: Postgres connection string (defaults to `DATABASE_URL`)
- `-stream`: UUID for this table’s stream id

Events are still recorded in-memory; Postgres writes are best-effort and failures are logged without interrupting the game loop.

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

## Environment Setup

### Installing Homebrew

`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

### Installing Go
The repository targets Go 1.22 and uses Go workspaces

```
brew install go
echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
go version
```

Synchronize workspace modules

`go work sync`

### Install Dev Tools

```
brew install golangci-lint flyway
golangci-lint --version
flyway -v
```

### Install Docker

```
brew install --cask docker
open -a Docker
docker --version
```

Start local infrastructure (Postgres, Redis, Kafka, Zookeeper, Consul)
Services are defined in `docker-compose.yml`

```
docker compose up -d
docker compose ps
```


