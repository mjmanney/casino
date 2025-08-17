SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := help
BIN_DIR := bin

.PHONY: help all go-build-services go-run-services go-run-all go-test go-lint go-clean docker-dev docker-down docker-ps docker-logs flyway-migrate flyway-clean

help: # Show targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' Makefile | sed 's/:.*##/: /'

all: docker-dev go-build-services ## Start infra & build

# --- Go / Services ---

go-build-services: ## Build api+game -> bin/
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/api  ./services/api
	go build -o $(BIN_DIR)/game ./services/game

go-run-services: # Run api+game via go run
	@echo "Starting API service..."
	go run ./services/api &
	API_PID=$$!
	@echo "Starting Game service..."
	go run ./services/game &
	GAME_PID=$$!
	trap "kill $$API_PID $$GAME_PID 2>/dev/null || true" INT TERM
	wait $$API_PID $$GAME_PID

go-run-all: go-build-services # Run compiled api+game from bin/
	./bin/api &
	API_PID=$$!
	./bin/game &
	GAME_PID=$$!
	trap "kill $$API_PID $$GAME_PID 2>/dev/null || true" INT TERM
	wait $$API_PID $$GAME_PID

go-test: # Run tests
	go test ./services/... ./libs/... -v

go-lint: # Lint (golangci-lint)
	golangci-lint run ./...

go-clean: # Clean artifacts
	rm -rf $(BIN_DIR)
	go clean ./...

# --- Docker / Infra ---

docker-dev:
	docker compose up -d

docker-down:
	docker compose down -v

docker-ps: # List containers
	@docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Ports}}"

docker-logs:
	docker compose logs -f

flyway-migrate:
	flyway migrate
flyway-clean:
	flyway clean
