# Problem Knowledge Base — dev helpers
# Usage: make <target>

.DEFAULT_GOAL := help
SHELL := /bin/bash

## ---- Docker ----------------------------------------------------
up: ## Start the full stack (postgres + backend + frontend)
	docker compose up -d --build

down: ## Stop the stack
	docker compose down

logs: ## Tail all container logs
	docker compose logs -f

nuke: ## Stop the stack and drop the database volume
	docker compose down -v

## ---- Backend (run on host) -----------------------------------
be-run: ## Run the API server locally (needs a reachable Postgres)
	cd backend && go run ./cmd/api

be-migrate: ## Apply database migrations
	cd backend && go run ./cmd/migrate up

be-seed: ## Insert seed data (categories, tags, admin, sample problem)
	cd backend && go run ./cmd/seed

be-build: ## Compile all backend packages
	cd backend && go build ./...

be-vet: ## go vet
	cd backend && go vet ./...

be-test: ## Run backend tests
	cd backend && go test ./...

## ---- Frontend (run on host) --------------------------------
fe-install: ## Install frontend dependencies
	cd frontend && npm install

fe-dev: ## Start the Vite dev server
	cd frontend && npm run dev

fe-build: ## Type-check and build the frontend
	cd frontend && npm run build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: up down logs nuke be-run be-migrate be-seed be-build be-vet be-test fe-install fe-dev fe-build help
