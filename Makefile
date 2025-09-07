SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := help
BIN_DIR := bin

.PHONY: help all go-build go-run go-test go-lint go-clean docker-dev docker-down docker-ps docker-logs flyway-migrate flyway-clean db-health db-events-count db-events-tail db-events-stream db-events-clear db-query

help: # Show targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' Makefile | sed 's/:.*##/: /'

all: ## Start infra & build
	@$(MAKE) docker-dev ; \
	$(MAKE) go-build-services

# --- Go / Services ---

go-build: ## Build to bin/
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/api  ./services/api
	go build -o $(BIN_DIR)/blackjack ./cmd/blackjack
	go build -o $(BIN_DIR)/tools ./cmd/tools/dbhealth

go-run:ß ## Run services
	@echo "Starting API service..." ; \
		go run ./services /api & \
		API_PID=$$! ; \
		echo "Starting Blackjack service (foreground)..." ; \
		trap "kill $$API_PID 2>/dev/null || true" INT TERM EXIT ; \
		go run ./cmd/blackjack ; \
		STATUS=$$? ; \
		kill $$API_PID 2>/dev/null || true ; \
		wait $$API_PID 2>/dev/null || true ; \
		exit $$STATUS
 
go-test: ## Run tests
	go test ./services/... ./libs/... -v

go-lint: ## Lint (golangci-lint)
	golangci-lint run ./...

go-clean: ## Clean artifacts
	rm -rf $(BIN_DIR)
	go clean -cache -testcache -modcache

# --- Docker / Infra ---

docker-dev:  ## Docker compose up
	export DATABASE_URL=postgres://postgres:postgres@localhost:5432/casino?sslmode=disable
	docker compose up -d

docker-down: ## Docker compose down
	docker compose down -v

docker-ps: ## List containers
	@docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Ports}}"

docker-logs: ## Docker logs
	docker compose logs -f

flyway-migrate: ## Flyway Migrate
	flyway migrate
flyway-clean: ## Flyway Clean
	flyway clean

# --- Postgres ---

db-health: ## Check Postgres connectivity using DATABASE_URL
	@if [ -z "$$DATABASE_URL" ]; then echo "Set DATABASE_URL or pass -db via MAKEFLAGS, e.g., make db-health MAKEFLAGS='-db=postgres://...'"; fi
	go run ./cmd/tools/dbhealth -db "$$DATABASE_URL"

db-events-count: ## Print total rows in event_log
	@docker exec casino_postgres psql -U postgres -d casino -t -A -c "SELECT COUNT(*) FROM event_log;" | sed 's/^/event_log count: /'

LIMIT ?= 10
db-events-tail: ## Print last ten events (override LIMIT=N)
	@docker exec casino_postgres psql -U postgres -d casino -P pager=off -c \
	"SELECT id, stream_id, seq, event_type, created_at FROM event_log ORDER BY id DESC LIMIT $(LIMIT);"

STREAM ?= 00000000-0000-0000-0000-000000000001
db-events-stream: ## Print events for STREAM (override STREAM=<uuid>, LIMIT=N)
	@docker exec casino_postgres psql -U postgres -d casino -P pager=off -c \
	"SELECT seq, event_type, created_at FROM event_log WHERE stream_id='$(STREAM)' ORDER BY seq DESC LIMIT $(LIMIT);"

db-events-clear: ## Truncate event_log (requires CONFIRM=delete)
	@if [ "$$CONFIRM" != "delete" ]; then \
		echo "Denied.  To proceed: make db-events-clear CONFIRM=delete"; \
		exit 1; \
	fi
	@docker exec casino_postgres psql -U postgres -d casino -P pager=off -c "TRUNCATE event_log RESTART IDENTITY;" && echo "event_log truncated."

QUERY ?= ""
db-query: ## Run a custom query against the database (override QUERY=<query>)
	@docker exec casino_postgres psql -U postgres -d casino -P pager=off -c \
	'$(QUERY)'
