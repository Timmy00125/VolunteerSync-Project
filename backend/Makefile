.PHONY: help build run test test-integration test-contract coverage lint fmt clean install-tools

# Default target
.DEFAULT_GOAL := help

# Colors for output
YELLOW := \033[1;33m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(YELLOW)VolunteerSync Backend - Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-18s$(NC) %s\n", $$1, $$2}'

build: ## Build the application
	@echo "$(YELLOW)Building application...$(NC)"
	go build -o bin/api cmd/api/main.go
	@echo "$(GREEN)✓ Build successful$(NC)"

run: ## Run the application
	@echo "$(YELLOW)Starting server...$(NC)"
	go run cmd/api/main.go

test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	go test ./... -v

test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	go test ./tests/integration/... -v

test-contract: ## Run contract tests
	@echo "$(YELLOW)Running contract tests...$(NC)"
	go test ./tests/contract/... -v

coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

lint: ## Run golangci-lint
	@echo "$(YELLOW)Running linter...$(NC)"
	@./scripts/lint.sh

fmt: ## Format code with gofmt and goimports
	@echo "$(YELLOW)Formatting code...$(NC)"
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
		echo "$(GREEN)✓ Code formatted with goimports$(NC)"; \
	else \
		echo "$(YELLOW)⚠ goimports not found, install with: go install golang.org/x/tools/cmd/goimports@latest$(NC)"; \
	fi

tidy: ## Tidy go.mod and go.sum
	@echo "$(YELLOW)Tidying dependencies...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

clean: ## Clean build artifacts and test cache
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache
	@echo "$(GREEN)✓ Cleaned$(NC)"

install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	@echo "Installing goimports..."
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Installing golangci-lint..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install golangci-lint; \
	else \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi
	@echo "Installing pre-commit (requires Python/pip)..."
	@if command -v pip3 > /dev/null; then \
		pip3 install pre-commit; \
	elif command -v pip > /dev/null; then \
		pip install pre-commit; \
	else \
		echo "$(YELLOW)⚠ pip not found, please install Python and pip to use pre-commit$(NC)"; \
	fi
	@echo "$(GREEN)✓ Development tools installed$(NC)"
	@echo "$(YELLOW)Setup pre-commit hooks with: pre-commit install$(NC)"

setup: install-tools tidy ## Setup development environment
	@echo "$(YELLOW)Setting up pre-commit hooks...$(NC)"
	@if command -v pre-commit > /dev/null; then \
		pre-commit install; \
		echo "$(GREEN)✓ Pre-commit hooks installed$(NC)"; \
	else \
		echo "$(YELLOW)⚠ pre-commit not found, run 'make install-tools' first$(NC)"; \
	fi
	@echo "$(GREEN)✓ Development environment ready!$(NC)"

all: fmt lint test ## Format, lint, and test

ci: tidy fmt lint test ## Run CI checks (format, lint, test)
	@echo "$(GREEN)✓ All CI checks passed$(NC)"
