# Makefile for MachinerySafety Backend
# =====================================

.PHONY: help build test run dev-up dev-up-full dev-down dev-build dev-build-backend \
        dev-build-frontend dev-logs dev-ps dev-clean dev-reset db-shell minio-console health version lint

# Default env file
ENV_FILE ?= .env

# Docker compose for local dev
DC_DEV = docker compose -f docker-compose.dev.yml

help: ## Show this help
	@echo "MachinerySafety - Development Commands"
	@echo "======================================="
	@echo ""
	@echo "Local Development (without Docker):"
	@echo "  make run           - Run server locally (requires running infra)"
	@echo "  make build         - Build Go binary"
	@echo "  make test          - Run all tests"
	@echo "  make lint          - Run golangci-lint"
	@echo ""
	@echo "Docker Development (full stack):"
	@echo "  make dev-up        - Start frontend + backend + postgres + minio"
	@echo "  make dev-up-full   - Start all services including metabase"
	@echo "  make dev-down      - Stop all services"
	@echo "  make dev-build     - Rebuild frontend and backend images"
	@echo "  make dev-build-backend  - Rebuild backend only"
	@echo "  make dev-build-frontend - Rebuild frontend only"
	@echo "  make dev-logs      - Tail all logs"
	@echo "  make dev-ps        - Show running containers"
	@echo "  make dev-clean     - Stop and remove all volumes (fresh start)"
	@echo "  make dev-reset     - Wipe volumes and restart with seeded DB"
	@echo ""
	@echo "Utilities:"
	@echo "  make health        - Check backend health endpoint"
	@echo "  make version       - Check backend version endpoint"
	@echo "  make db-shell      - Open psql shell"
	@echo "  make minio-console - Open minio console URL"
	@echo ""

# =============================================================================
# Local Development (without Docker)
# =============================================================================

build: ## Build Go binary
	go build -ldflags "-X github.com/zserub/machinerysafety-BE/internal/handler.Version=dev-local" -o bin/server .

test: ## Run all tests
	go test -race -v ./...

test-short: ## Run tests (skip slow tests)
	go test -short -v ./...

test-cover: ## Run tests with coverage
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run: ## Run server locally (requires DB, MinIO to be running)
	@if [ -f $(ENV_FILE) ]; then \
		export $$(grep -v '^#' $(ENV_FILE) | xargs) && go run main.go; \
	else \
		echo "No $(ENV_FILE) found. Copy .env.example to .env and configure."; \
		exit 1; \
	fi

lint: ## Run golangci-lint
	golangci-lint run ./...

# =============================================================================
# Docker Development (full stack)
# =============================================================================

dev-up: ## Start development stack (frontend + backend + postgres + minio)
	$(DC_DEV) up -d
	@echo ""
	@echo "Services starting..."
	@echo "  Frontend: http://localhost:3001"
	@echo "  Backend:  http://localhost:8080"
	@echo "  MinIO:    http://localhost:9001 (console)"
	@echo ""
	@echo "Run 'make dev-logs' to tail logs"
	@echo "Run 'make health' to check backend status"

dev-up-full: ## Start all services including metabase (auto-creates dashboard if METABASE_API_KEY is set)
	RUN_METABASE_SETUP=true $(DC_DEV) --profile metabase up -d
	@echo ""
	@echo "Services starting..."
	@echo "  Frontend:  http://localhost:3001"
	@echo "  Backend:   http://localhost:8080"
	@echo "  MinIO:     http://localhost:9001 (console)"
	@echo "  Metabase:  http://localhost:3000"
	@echo ""

dev-down: ## Stop all services
	$(DC_DEV) --profile metabase down

dev-build: ## Rebuild frontend and backend images and restart (migrations auto-applied on backend start)
	$(DC_DEV) build frontend backend
	$(DC_DEV) up -d frontend backend

dev-reset: ## Fresh start: wipe all volumes and reseed DB from scratch
	$(DC_DEV) --profile metabase down -v
	$(DC_DEV) up -d
	@echo ""
	@echo "Fresh environment starting with seeded database..."
	@echo "Run 'make dev-logs-backend' to watch migration progress"

dev-build-backend: ## Rebuild backend image only and restart
	$(DC_DEV) build backend
	$(DC_DEV) up -d backend

dev-build-frontend: ## Rebuild frontend image only and restart
	$(DC_DEV) build frontend
	$(DC_DEV) up -d frontend

dev-restart: ## Restart backend service
	$(DC_DEV) restart backend

dev-restart-frontend: ## Restart frontend service
	$(DC_DEV) restart frontend

dev-logs: ## Tail all logs
	$(DC_DEV) logs -f

dev-logs-backend: ## Tail backend logs only
	$(DC_DEV) logs -f backend

dev-logs-frontend: ## Tail frontend logs only
	$(DC_DEV) logs -f frontend

dev-ps: ## Show running containers
	$(DC_DEV) ps

dev-clean: ## Stop and remove all volumes (fresh start)
	$(DC_DEV) --profile metabase down -v
	@echo "All containers and volumes removed. Run 'make dev-up' for fresh start."

dev-shell: ## Open shell in backend container
	$(DC_DEV) exec backend /bin/sh

dev-shell-frontend: ## Open shell in frontend container
	$(DC_DEV) exec frontend /bin/sh

# =============================================================================
# Utilities
# =============================================================================

health: ## Check backend health endpoint
	@curl -s http://localhost:8080/health/live | jq . 2>/dev/null || curl -s http://localhost:8080/health/live

version: ## Check backend version endpoint
	@curl -s http://localhost:8080/api/version | jq . 2>/dev/null || curl -s http://localhost:8080/api/version

db-shell: ## Open psql shell in postgres container
	$(DC_DEV) exec postgres psql -U postgres -d machinery_safety

db-logs: ## Show postgres logs
	$(DC_DEV) logs -f postgres

minio-console: ## Print MinIO console URL
	@echo "MinIO Console: http://localhost:9001"
	@echo "Credentials: minioadmin / password (or check .env)"

# =============================================================================
# CI/Release helpers
# =============================================================================

docker-build: ## Build production docker image
	docker build \
		--build-arg VERSION=$$(git describe --tags --always --dirty) \
		-t machinery-safety-backend:$$(git describe --tags --always --dirty) \
		.

docker-build-dev: ## Build docker image with dev tag
	docker build \
		--build-arg VERSION=dev-$$(git rev-parse --short HEAD) \
		-t machinery-safety-backend:dev \
		.
