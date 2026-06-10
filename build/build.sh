#!/bin/bash

# chklst-go Build Script
# Builds Go binary with embedded Python microservice

set -e

echo "🚀 chklst-go Build Script"
echo "=========================="

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"
PYTHON_SERVICE_DIR="$PROJECT_ROOT/python-service"
GO_MAIN="$PROJECT_ROOT/cmd/chklst/main.go"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
print_step() {
    echo -e "${BLUE}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Clean previous builds
print_step "Cleaning previous builds..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
print_success "Clean complete"

# (Step 1 removed: the optional Python microservice no longer exists.)

# Step 2: Build Go Binary
print_step "Building Go binary..."

cd "$PROJECT_ROOT"

# Get version from git or use default
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -s -w"

# Build for current platform
print_step "Building for current platform..."
go build -ldflags "$LDFLAGS" -o "$DIST_DIR/chklst" "$GO_MAIN"
print_success "Go binary built: $DIST_DIR/chklst"

# Cross-compile for other platforms (optional)
if [ "$BUILD_ALL_PLATFORMS" = "true" ]; then
    print_step "Cross-compiling for multiple platforms..."

    # Linux AMD64
    GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" \
        -o "$DIST_DIR/chklst-linux-amd64" "$GO_MAIN"
    print_success "Built for Linux AMD64"

    # Linux ARM64
    GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" \
        -o "$DIST_DIR/chklst-linux-arm64" "$GO_MAIN"
    print_success "Built for Linux ARM64"

    # macOS AMD64
    GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" \
        -o "$DIST_DIR/chklst-darwin-amd64" "$GO_MAIN"
    print_success "Built for macOS AMD64"

    # macOS ARM64 (Apple Silicon)
    GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" \
        -o "$DIST_DIR/chklst-darwin-arm64" "$GO_MAIN"
    print_success "Built for macOS ARM64"

    # Windows AMD64
    GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" \
        -o "$DIST_DIR/chklst-windows-amd64.exe" "$GO_MAIN"
    print_success "Built for Windows AMD64"
fi

# (Python microservice removed — exports are handled in the frontend/Go directly.)

# Step 4: Create deployment package
print_step "Creating deployment package..."

# Create README for deployment
cat > "$DIST_DIR/README.txt" << 'EOF'
chklst-go Deployment
====================

Quick Start:
1. Run the binary: ./chklst
2. Access at http://localhost:8000
3. Health check: http://localhost:8000/health

Python Service (if standalone):
1. cd python-service
2. pip install -r requirements.txt
3. python main.py

Configuration:
- DB_PATH: Database file path (default: ./chklst.db)
- BACKUP_DIR: Backup directory (default: ./backups)
- PORT: Server port (default: 8000)
- AUTO_BACKUP_HOURS: Auto-backup interval (default: 24)

Copy your existing chklst.db to the same directory as the binary.

For more information, see README.md in the source repository.
EOF

print_success "Deployment package created"

# Summary
echo ""
echo "=========================="
print_success "Build Complete!"
echo "=========================="
echo ""
echo "Build artifacts:"
ls -lh "$DIST_DIR"
echo ""
echo "Usage:"
echo "  cd $DIST_DIR"
echo "  ./chklst"
echo ""

if [ "$BUILD_ALL_PLATFORMS" = "true" ]; then
    echo "Cross-compiled binaries available for:"
    echo "  - Linux AMD64/ARM64"
    echo "  - macOS AMD64/ARM64 (Intel/Apple Silicon)"
    echo "  - Windows AMD64"
    echo ""
fi

print_success "Ready to deploy! 🚀"
