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

# Docker-based multi-version testing
GO_VERSIONS := 1.20 1.21 1.22 1.23
DOCKER_RUN := docker run --rm -v "$(PWD)":/workspace -w /workspace

docker-test: ## Test with all supported Go versions using Docker
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)Testing with Multiple Go Versions$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@for version in $(GO_VERSIONS); do \
		echo ""; \
		echo "$(YELLOW)Testing with Go $$version...$(NC)"; \
		echo "----------------------------------------"; \
		$(DOCKER_RUN) golang:$$version sh -c "go version && go test -v ./..." || exit 1; \
		echo "$(GREEN)✓ Go $$version tests passed$(NC)"; \
	done
	@echo ""
	@echo "$(GREEN)All version tests completed successfully!$(NC)"

docker-test-version: ## Test with specific Go version (use GO_VERSION=1.20)
	@if [ -z "$(GO_VERSION)" ]; then \
		echo "$(RED)Please specify GO_VERSION, e.g., make docker-test-version GO_VERSION=1.20$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Testing with Go $(GO_VERSION)...$(NC)"
	@$(DOCKER_RUN) golang:$(GO_VERSION) sh -c "go version && go test -v -race -cover ./..."

docker-test-matrix: ## Run comprehensive test matrix with all Go versions
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)Running Test Matrix$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@for version in $(GO_VERSIONS); do \
		echo ""; \
		echo "$(YELLOW)═══════════════════════════════════════$(NC)"; \
		echo "$(YELLOW)   Go $$version Test Suite$(NC)"; \
		echo "$(YELLOW)═══════════════════════════════════════$(NC)"; \
		echo ""; \
		echo "$(YELLOW)[1/4] Basic Tests$(NC)"; \
		$(DOCKER_RUN) golang:$$version go test -short ./... || exit 1; \
		echo "$(GREEN)✓ Basic tests passed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)[2/4] Race Detection$(NC)"; \
		$(DOCKER_RUN) golang:$$version go test -race ./... || exit 1; \
		echo "$(GREEN)✓ Race detection passed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)[3/4] Coverage Analysis$(NC)"; \
		$(DOCKER_RUN) golang:$$version sh -c "go test -cover ./... | grep -E 'coverage:|ok'" || exit 1; \
		echo "$(GREEN)✓ Coverage analysis completed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)[4/4] Build Verification$(NC)"; \
		$(DOCKER_RUN) golang:$$version go build -v ./... || exit 1; \
		echo "$(GREEN)✓ Build verification passed$(NC)"; \
		echo ""; \
		echo "$(GREEN)═══════════════════════════════════════$(NC)"; \
		echo "$(GREEN)✓ Go $$version: ALL TESTS PASSED$(NC)"; \
		echo "$(GREEN)═══════════════════════════════════════$(NC)"; \
	done
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ Test Matrix Completed Successfully!$(NC)"
	@echo "$(GREEN)========================================$(NC)"

docker-bench: ## Run benchmarks with all Go versions using Docker
	@echo "$(GREEN)Running benchmarks across Go versions...$(NC)"
	@for version in $(GO_VERSIONS); do \
		echo ""; \
		echo "$(YELLOW)Benchmarks with Go $$version:$(NC)"; \
		$(DOCKER_RUN) golang:$$version go test -bench=. -benchmem ./... | grep -E "Benchmark|ns/op|allocs/op"; \
	done

docker-coverage: ## Generate coverage report using Docker
	@echo "$(GREEN)Generating coverage report with Docker...$(NC)"
	@$(DOCKER_RUN) golang:1.22 sh -c "\
		go test -v -coverprofile=coverage.out ./... && \
		go tool cover -func=coverage.out && \
		echo '' && \
		echo 'Total Coverage:' && \
		go tool cover -func=coverage.out | grep total"

docker-lint: ## Run linting with golangci-lint in Docker
	@echo "$(GREEN)Running linters in Docker...$(NC)"
	@docker run --rm -v "$(PWD)":/workspace -w /workspace golangci/golangci-lint:latest \
		golangci-lint run --timeout=5m ./...

docker-clean: ## Clean Docker test artifacts and containers
	@echo "$(GREEN)Cleaning Docker artifacts...$(NC)"
	@docker system prune -f --volumes --filter "label=test=golang-yaml-advanced" 2>/dev/null || true
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Docker cleanup complete$(NC)"

docker-shell: ## Open interactive shell in Docker container with Go
	@echo "$(GREEN)Opening Docker shell with Go $(or $(GO_VERSION),1.22)...$(NC)"
	@$(DOCKER_RUN) -it golang:$(or $(GO_VERSION),1.22) /bin/bash

docker-test-quick: ## Quick test with latest 3 Go versions
	@echo "$(GREEN)Quick testing with Go 1.21, 1.22, 1.23...$(NC)"
	@for version in 1.21 1.22 1.23; do \
		echo "$(YELLOW)Testing Go $$version...$(NC)"; \
		$(DOCKER_RUN) golang:$$version go test ./... || exit 1; \
		echo "$(GREEN)✓$(NC)"; \
	done

docker-test-compat: ## Test backwards compatibility (Go 1.20)
	@echo "$(GREEN)Testing backwards compatibility with Go 1.20...$(NC)"
	@$(DOCKER_RUN) golang:1.20 sh -c "\
		echo 'Go version:' && go version && \
		echo '' && \
		echo 'Testing...' && \
		go test -v ./..."

docker-test-latest: ## Test with latest Go version
	@echo "$(GREEN)Testing with latest Go version...$(NC)"
	@$(DOCKER_RUN) golang:latest sh -c "\
		echo 'Go version:' && go version && \
		echo '' && \
		go test -v -race -cover ./..."

docker-ci: ## Run full CI pipeline in Docker
	@echo "$(GREEN)Running CI pipeline in Docker...$(NC)"
	@$(DOCKER_RUN) golang:1.22 sh -c "\
		echo '=== Downloading dependencies ===' && \
		go mod download && \
		echo '' && \
		echo '=== Running go vet ===' && \
		go vet ./... && \
		echo '' && \
		echo '=== Running tests ===' && \
		go test -v -race -cover ./... && \
		echo '' && \
		echo '=== Building ===' && \
		go build -v ./..."

# Docker Compose based testing
compose-test: ## Run tests with all Go versions using Docker Compose
	@echo "$(GREEN)Running tests with Docker Compose...$(NC)"
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-go120 test-go120
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-go121 test-go121
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-go122 test-go122
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test-go123 test-go123

compose-test-parallel: ## Run all tests in parallel using Docker Compose
	@echo "$(GREEN)Running all tests in parallel...$(NC)"
	@docker-compose -f docker-compose.test.yml up \
		test-go120 test-go121 test-go122 test-go123 test-latest

compose-test-single: ## Run test for single Go version (use SERVICE=test-go120)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Please specify SERVICE, e.g., make compose-test-single SERVICE=test-go120$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running $(SERVICE)...$(NC)"
	@docker-compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from $(SERVICE) $(SERVICE)

compose-bench: ## Run benchmarks using Docker Compose
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@docker-compose -f docker-compose.test.yml up benchmark

compose-coverage: ## Generate coverage report using Docker Compose
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@mkdir -p coverage
	@docker-compose -f docker-compose.test.yml up coverage
	@echo "$(GREEN)Coverage report saved to coverage/$(NC)"

compose-lint: ## Run linting using Docker Compose
	@echo "$(GREEN)Running linters...$(NC)"
	@docker-compose -f docker-compose.test.yml up lint

compose-security: ## Run security scan using Docker Compose
	@echo "$(GREEN)Running security scan...$(NC)"
	@docker-compose -f docker-compose.test.yml up security

compose-all: ## Run all tests, benchmarks, coverage, and linting
	@echo "$(GREEN)Running complete test suite with Docker Compose...$(NC)"
	@docker-compose -f docker-compose.test.yml up

compose-down: ## Stop and remove all test containers
	@echo "$(GREEN)Cleaning up Docker Compose containers...$(NC)"
	@docker-compose -f docker-compose.test.yml down -v

compose-logs: ## Show logs from all test containers
	@docker-compose -f docker-compose.test.yml logs

compose-ps: ## Show status of test containers
	@docker-compose -f docker-compose.test.yml ps

# Help for Docker targets
docker-help: ## Show Docker-specific targets
	@echo "$(GREEN)Docker Testing Targets:$(NC)"
	@echo ""
	@echo "$(YELLOW)── Docker Run Commands ──$(NC)"
	@echo "  $(YELLOW)docker-test$(NC)          - Test with all supported Go versions"
	@echo "  $(YELLOW)docker-test-version$(NC)  - Test with specific version (GO_VERSION=1.20)"
	@echo "  $(YELLOW)docker-test-matrix$(NC)   - Comprehensive test matrix"
	@echo "  $(YELLOW)docker-test-quick$(NC)    - Quick test with recent versions"
	@echo "  $(YELLOW)docker-test-compat$(NC)   - Test Go 1.20 compatibility"
	@echo "  $(YELLOW)docker-test-latest$(NC)   - Test with latest Go"
	@echo "  $(YELLOW)docker-bench$(NC)         - Run benchmarks across versions"
	@echo "  $(YELLOW)docker-coverage$(NC)      - Generate coverage report"
	@echo "  $(YELLOW)docker-lint$(NC)          - Run golangci-lint"
	@echo "  $(YELLOW)docker-ci$(NC)            - Run full CI pipeline"
	@echo "  $(YELLOW)docker-shell$(NC)         - Interactive Docker shell"
	@echo "  $(YELLOW)docker-clean$(NC)         - Clean Docker artifacts"
	@echo ""
	@echo "$(YELLOW)── Docker Compose Commands ──$(NC)"
	@echo "  $(YELLOW)compose-test$(NC)         - Sequential tests with all versions"
	@echo "  $(YELLOW)compose-test-parallel$(NC) - Parallel tests with all versions"
	@echo "  $(YELLOW)compose-test-single$(NC)  - Test single version (SERVICE=test-go120)"
	@echo "  $(YELLOW)compose-bench$(NC)        - Run benchmarks"
	@echo "  $(YELLOW)compose-coverage$(NC)     - Generate coverage report"
	@echo "  $(YELLOW)compose-lint$(NC)         - Run linting"
	@echo "  $(YELLOW)compose-security$(NC)     - Run security scan"
	@echo "  $(YELLOW)compose-all$(NC)          - Run everything"
	@echo "  $(YELLOW)compose-down$(NC)         - Clean up containers"
	@echo ""
	@echo "$(GREEN)Examples:$(NC)"
	@echo "  make docker-test"
	@echo "  make docker-test-version GO_VERSION=1.21"
	@echo "  make compose-test-parallel"
	@echo "  make compose-test-single SERVICE=test-go122"
	@echo "  make docker-shell GO_VERSION=1.20"