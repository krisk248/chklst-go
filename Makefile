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

install-python: ## Install Python dependencies
	@echo "🐍 Installing Python dependencies..."
	@cd python-service && pip install -r requirements.txt
	@echo "✅ Python dependencies installed"

dev: ## Quick development build
	@./build/dev-build.sh

build: ## Full production build
	@./build/build.sh

build-all: ## Build for all platforms
	@BUILD_ALL_PLATFORMS=true ./build/build.sh

build-embedded: ## Build with embedded Python service
	@BUILD_PYTHON=true ./build/build.sh

run: ## Run the application
	@echo "🚀 Starting chklst-go..."
	@go run cmd/chklst/main.go

run-python: ## Run Python microservice standalone
	@echo "🐍 Starting Python reports service..."
	@cd python-service && python main.py

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

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t chklst-go:latest .
	@echo "✅ Docker image built"

docker-run: ## Run in Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8000:8000 -v $(PWD)/chklst.db:/app/chklst.db chklst-go:latest

# Default target
.DEFAULT_GOAL := help
