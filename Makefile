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

# Default target
.DEFAULT_GOAL := help
