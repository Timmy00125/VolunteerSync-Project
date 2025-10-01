#!/bin/bash
# lint.sh - Run golangci-lint on the VolunteerSync backend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running golangci-lint...${NC}"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}golangci-lint is not installed.${NC}"
    echo "Install it with:"
    echo "  brew install golangci-lint  # macOS"
    echo "  or visit: https://golangci-lint.run/usage/install/"
    exit 1
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

cd "$BACKEND_DIR"

# Run golangci-lint
if golangci-lint run --config .golangci.yml ./...; then
    echo -e "${GREEN}✓ Linting passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Linting failed. Please fix the issues above.${NC}"
    exit 1
fi
