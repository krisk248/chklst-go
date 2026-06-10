#!/bin/bash

# Quick Development Build Script
# Builds Go binary only for rapid testing

set -e

echo "🔧 Development Build"
echo "===================="

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_MAIN="$PROJECT_ROOT/cmd/chklst/main.go"

cd "$PROJECT_ROOT"

echo "▶ Building Go binary..."
go build -o chklst "$GO_MAIN"

echo "✅ Build complete!"
echo ""
echo "Run with: ./chklst"
echo "API: http://localhost:8000"
echo "Health: http://localhost:8000/health"
echo ""
echo "Environment variables:"
echo "  export DB_PATH=./chklst.db"
echo "  export PORT=8000"
