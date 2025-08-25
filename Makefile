SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := help
BIN_DIR := bin

.PHONY: help all go-build-services go-run-services go-run-all go-test go-lint go-clean docker-dev docker-down docker-ps docker-logs flyway-migrate flyway-clean

help: # Show targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' Makefile | sed 's/:.*##/: /'

all: ## Start infra & build
	docker-dev 
	go-build-services

# --- Go / Services ---

go-build-services: ## Build api+game -> bin/
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/api  ./services/api
	go build -o $(BIN_DIR)/blackjack ./services/blackjack

go-run-services: ## Run api+game via go run
	@echo "Starting API service..."
	go run ./services/api &
	API_PID=$$!
	@echo "Starting Blackjack service..."
	go run ./services/blackjack &
	GAME_PID=$$!
	trap "kill $$API_PID $$GAME_PID 2>/dev/null || true" INT TERM
	wait $$API_PID $$GAME_PID

go-run-all: ## Run compiled api+game from bin/
	go-build-services
	./bin/api &
	API_PID=$$!
	./bin/blackjack &
	GAME_PID=$$!
	trap "kill $$API_PID $$GAME_PID 2>/dev/null || true" INT TERM
	wait $$API_PID $$GAME_PID

go-test: ## Run tests
	go test ./services/... ./libs/... -v

go-lint: ## Lint (golangci-lint)
	golangci-lint run ./...

go-clean: ## Clean artifacts
	rm -rf $(BIN_DIR)
	go clean ./...

# --- Docker / Infra ---

docker-dev:  ## Docker compose up
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
