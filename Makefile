.PHONY: test coverage bench clean lint install help test-complete

# Variables
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html
PKG_PATH := ./...
PKG_TEST_PATH := ./
EXAMPLES_PATH := ./examples/

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

# Default target
help: ## Show this help message
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

deps: ## Install deps package
	@echo "$(GREEN)Install shadow vettool package...$(NC)"
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod download
	go mod tidy

test: ## Run all tests
	@echo "$(GREEN)Running tests for yaml package...$(NC)"
	go test -v $(PKG_TEST_PATH)

coverage: ## Run tests with coverage report
	@echo "$(GREEN)Running unit tests with coverage...$(NC)"
	go test -v -cover -coverprofile=$(COVERAGE_OUT) $(PKG_TEST_PATH)
	@echo ""
	@echo "$(GREEN)Coverage Report:$(NC)"
	go tool cover -func=$(COVERAGE_OUT)
	@echo ""
	@echo "$(YELLOW)To view HTML coverage report, run: make coverage-html$(NC)"

coverage-html: ## Generate and open HTML coverage report
	@echo "$(GREEN)Generating HTML coverage report...$(NC)"
	go test -v -cover -coverprofile=$(COVERAGE_OUT) $(PKG_TEST_PATH)
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report saved to $(COVERAGE_HTML)$(NC)"
	@if command -v open > /dev/null; then \
		open $(COVERAGE_HTML); \
	elif command -v xdg-open > /dev/null; then \
		xdg-open $(COVERAGE_HTML); \
	else \
		echo "$(YELLOW)Please open $(COVERAGE_HTML) in your browser$(NC)"; \
	fi

bench: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem $(PKG_TEST_PATH)

test-race: ## Run tests with race detector
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	go test -race -v $(PKG_TEST_PATH)

test-short: ## Run tests in short mode (skip long tests)
	@echo "$(GREEN)Running tests in short mode...$(NC)"
	go test -short -v $(PKG_TEST_PATH)

test-all: ## Run all tests including benchmarks
	@echo "$(GREEN)Running complete test suite...$(NC)"
	go test -v -cover -race -bench=. $(PKG_TEST_PATH)

test-complete: ## Complete test suite with coverage, benchmarks, and HTML report (equivalent to run_tests.sh)
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)Running Complete Test Suite$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Step 1: Running unit tests with coverage...$(NC)"
	go test -v -cover -coverprofile=$(COVERAGE_OUT) $(PKG_TEST_PATH)
	@echo ""
	@echo "$(YELLOW)Step 2: Coverage Report:$(NC)"
	go tool cover -func=$(COVERAGE_OUT)
	@echo ""
	@echo "$(YELLOW)Step 3: Generating HTML coverage report...$(NC)"
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report saved to $(COVERAGE_HTML)$(NC)"
	@echo ""
	@echo "$(YELLOW)Step 4: Running benchmarks...$(NC)"
	go test -bench=. -benchmem $(PKG_TEST_PATH)
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)All tests completed successfully!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Summary:$(NC)"
	@echo "  - Unit tests: ✓"
	@echo "  - Coverage report: $(COVERAGE_OUT)"
	@echo "  - HTML report: $(COVERAGE_HTML)"
	@echo "  - Benchmarks: ✓"

test-verbose: ## Run tests with detailed verbose output
	@echo "$(GREEN)Running tests with verbose output...$(NC)"
	go test -v -count=1 $(PKG_TEST_PATH)

test-json: ## Run tests with JSON output (for CI/CD integration)
	@echo "$(GREEN)Running tests with JSON output...$(NC)"
	go test -json $(PKG_TEST_PATH)

coverage-threshold: ## Check if coverage meets minimum threshold (80%)
	@echo "$(GREEN)Checking coverage threshold...$(NC)"
	@go test -cover -coverprofile=$(COVERAGE_OUT) $(PKG_TEST_PATH) > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_OUT) | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= 80" | bc) -eq 1 ]; then \
		echo "$(GREEN)✓ Coverage meets minimum threshold (80%)$(NC)"; \
	else \
		echo "$(RED)✗ Coverage below minimum threshold (80%)$(NC)"; \
		exit 1; \
	fi

clean: ## Clean build and test artifacts
	@echo "$(GREEN)Cleaning test artifacts...$(NC)"
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	go clean -testcache
	@echo "$(GREEN)Clean complete$(NC)"

lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	@which golangci-lint > /dev/null 2>&1 || (echo "$(YELLOW)Installing golangci-lint...$(NC)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt $(PKG_PATH)
	gofmt -s -w .

vet: deps ## Run go vet
	@echo "$(GREEN)Running vet for yaml package...$(NC)"
	go vet -vettool=$(which shadow) $(PKG_PATH)
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet $(PKG_PATH)

build: ## Build the library
	@echo "$(GREEN)Building library...$(NC)"
	go build $(PKG_TEST_PATH)

examples: ## Run example programs
	@echo "$(GREEN)Running demo example...$(NC)"
	go run $(EXAMPLES_PATH)demo/main.go
	@echo ""
	@echo "$(GREEN)Running advanced demo...$(NC)"
	go run $(EXAMPLES_PATH)advanced_demo/main.go

examples-clean: ## Clean example output files
	@echo "$(GREEN)Cleaning example output files...$(NC)"
	rm -f output_*.yaml
	@echo "$(GREEN)Example outputs cleaned$(NC)"

ci: lint vet test coverage coverage-threshold ## Run CI pipeline (lint, vet, test, coverage with threshold)
	@echo "$(GREEN)CI pipeline completed successfully$(NC)"

# Development targets
dev-test: ## Run tests with verbose output for development
	@echo "$(GREEN)Running development tests...$(NC)"
	go test -v -count=1 $(PKG_TEST_PATH)

watch: ## Watch for changes and run tests (requires entr)
	@which entr > /dev/null 2>&1 || (echo "$(RED)Please install 'entr' to use watch mode$(NC)" && exit 1)
	@echo "$(GREEN)Watching for changes...$(NC)"
	find . -name '*.go' | entr -c go test -v $(PKG_TEST_PATH)

# Quick commands
quick: test ## Quick test run (alias for test)

full: test-complete ## Full test suite (alias for test-complete)

# Report generation
reports: coverage-html ## Generate all reports
	@echo "$(GREEN)All reports generated$(NC)"

# Installation verification
verify: install test ## Verify installation and run basic tests
	@echo "$(GREEN)Verification complete$(NC)"