# Makefile for chklst-go

.PHONY: help build dev run test clean install deps

help: ## Show this help message
	@echo "chklst-go - Deployment Checklist Tool"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

deps: ## Install Go dependencies
	@echo "📦 Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

dev: ## Quick development build
	@./build/dev-build.sh

build: ## Full production build
	@./build/build.sh

build-all: ## Build for all platforms
	@BUILD_ALL_PLATFORMS=true ./build/build.sh

run: ## Run the application
	@echo "🚀 Starting chklst-go..."
	@go run cmd/chklst/main.go

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

lint: ## Run linters
	@echo "🔍 Running linters..."
	@go vet ./...
	@echo "✅ Linting complete"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf dist/
	@rm -rf build/pyinstaller-work/
	@rm -rf build/*.spec
	@rm -f chklst
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

format: ## Format code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Formatting complete"

migrate-db: ## Copy existing database
	@echo "📊 Copying existing database..."
	@cp ../chklst.db ./chklst.db 2>/dev/null || echo "Note: ../chklst.db not found"
	@echo "✅ Database ready"

# --- Docker Compose workflow ---
# A clean throwaway docker config dir avoids the host's `credsStore` credential-helper
# error during registry pulls. (Permanent fix: remove "credsStore" from ~/.docker/config.json.)
DOCKER_CFG := /tmp/chklst-dockercfg
COMPOSE := DOCKER_CONFIG=$(DOCKER_CFG) docker compose

$(DOCKER_CFG)/config.json:
	@mkdir -p $(DOCKER_CFG) && printf '{}' > $(DOCKER_CFG)/config.json

up redeploy: $(DOCKER_CFG)/config.json ## Rebuild image from Dockerfile and (re)start on :8000
	@echo "🐳 Rebuilding + starting chklst..."
	@$(COMPOSE) up -d --build
	@echo "✅ chklst running at http://localhost:8000"

down: ## Stop and remove the container (keeps data volumes)
	@$(COMPOSE) down

logs: ## Follow container logs (Parson scheduler, requests)
	@$(COMPOSE) logs -f

ps: ## Show container status
	@$(COMPOSE) ps

docker-build: $(DOCKER_CFG)/config.json ## Build the image only
	@$(COMPOSE) build

# --- Parallel DEV instance on :9000 (in-progress branch, copy of real data) ---
DEV_COMPOSE := DOCKER_CONFIG=$(DOCKER_CFG) docker compose -f docker-compose.dev.yml

dev-seed: $(DOCKER_CFG)/config.json ## Copy the live :8000 DB into the dev volume (snapshot)
	@$(DEV_COMPOSE) up -d --build >/dev/null 2>&1 || true
	@$(DEV_COMPOSE) stop >/dev/null 2>&1 || true
	@docker run --rm -v chklst-go_chklst-data:/src -v chklst-go_chklst-dev-data:/dst alpine \
		sh -c "cp -f /src/chklst.db /dst/chklst.db 2>/dev/null && echo seeded || echo 'no source DB yet'"

dev-up: $(DOCKER_CFG)/config.json ## Rebuild + start the dev instance on :9000
	@echo "🧪 Starting dev (branch) on :9000 ..."
	@$(DEV_COMPOSE) up -d --build
	@echo "✅ dev at http://localhost:9000 (scheduler disabled, isolated data)"

dev-down: ## Stop the dev instance (keeps dev data volume)
	@$(DEV_COMPOSE) down

dev-logs: ## Follow dev instance logs
	@$(DEV_COMPOSE) logs -f

# Default target
.DEFAULT_GOAL := help
