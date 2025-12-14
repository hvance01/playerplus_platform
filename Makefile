.PHONY: dev dev-backend dev-frontend build clean db-up db-down db-migrate

# Development
dev:
	@echo "Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	@export $$(cat .env | xargs) && cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && pnpm dev

# Build
build: build-frontend build-backend

build-frontend:
	cd frontend && pnpm install && pnpm build
	@mkdir -p backend/internal/handler/dist
	@cp -r frontend/dist/* backend/internal/handler/dist/

build-backend: build-frontend
	cd backend && go build -o bin/server ./cmd/server

# Production run (local test)
run: build
	@export $$(cat .env | xargs) && ./backend/bin/server

# Clean
clean:
	rm -rf backend/bin
	rm -rf backend/internal/handler/dist
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Dependencies
deps:
	cd backend && go mod tidy
	cd frontend && pnpm install

# Lint
lint:
	cd backend && golangci-lint run
	cd frontend && pnpm lint

# Test
test:
	cd backend && go test ./...
	cd frontend && pnpm test

# Database (Docker)
db-up:
	docker compose up -d postgres
	@echo "PostgreSQL running on localhost:5432"
	@echo "Connection: postgresql://postgres:postgres@localhost:5432/playplus"

db-down:
	docker compose down

db-migrate:
	@export $$(cat .env | xargs) && psql $$DATABASE_URL -f backend/migrations/001_init.sql

db-logs:
	docker compose logs -f postgres
