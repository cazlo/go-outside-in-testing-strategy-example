.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Development

.PHONY: deps
deps: ## Download Go module dependencies
	go mod download

.PHONY: tidy
tidy: ## Tidy Go module dependencies
	go mod tidy

.PHONY: build
build: ## Build the server binary locally
	go build -o bin/server ./cmd/server

.PHONY: run
run: ## Run the server locally (default config)
	go run ./cmd/server

.PHONY: run-with-mocks
run-with-mocks: deps-up ## Run the server locally with wiremock dependency
	EXTERNAL_URL=http://localhost:8081/status/204 go run ./cmd/server

##@ Testing

.PHONY: test
test: ## Run unit tests only (use test-integration for full outside-in tests)
	go test ./internal/... -count=1 -v

.PHONY: test-unit
test-unit: test ## Alias for test (unit tests only)

.PHONY: test-blackbox
test-blackbox: ## Run blackbox tests against a running server (requires BASE_URL)
	go test ./test/blackbox -count=1 -v

.PHONY: test-blackbox-local
test-blackbox-local: ## Run blackbox tests against local server (assumes server running on :8080)
	BASE_URL=http://localhost:8080 go test ./test/blackbox -count=1 -v

.PHONY: test-integration
test-integration: deps-up ## Run integration tests with dependencies (server must be running separately)
	@echo "Starting wiremock dependency..."
	@echo "Run 'make run-with-mocks' in another terminal, then 'make test-blackbox-local'"
	@echo "Or use 'make test-integration-with-coverage' for automated testing with coverage"

.PHONY: build-coverage
build-coverage: ## Build server binary with coverage instrumentation
	go build -cover -o bin/server-coverage ./cmd/server

.PHONY: test-integration-with-coverage
test-integration-with-coverage: deps-up build-coverage ## Run integration tests with coverage collection
	@echo "Starting coverage-instrumented server..."
	@mkdir -p coverage
	@rm -rf coverage/cov* 2>/dev/null || true
	@GOCOVERDIR=coverage EXTERNAL_URL=http://localhost:8081/status/204 ./bin/server-coverage & echo $$! > .server.pid
	@sleep 2
	@echo "Running blackbox tests..."
	@BASE_URL=http://localhost:8080 go test ./test/blackbox -count=1 -v || (kill -SIGTERM `cat .server.pid` 2>/dev/null; sleep 1; rm -f .server.pid; exit 1)
	@echo "Stopping server gracefully..."
	@kill -SIGTERM `cat .server.pid` 2>/dev/null || true
	@wait `cat .server.pid` 2>/dev/null || true
	@rm -f .server.pid
	@echo "Converting coverage data..."
	@if [ -f coverage/covcounters.* ]; then \
		go tool covdata textfmt -i=coverage -o coverage/coverage.out; \
		go tool cover -html=coverage/coverage.out -o coverage/coverage.html; \
		echo "Coverage report generated: coverage/coverage.html"; \
		go tool cover -func=coverage/coverage.out | tail -n 1; \
	else \
		echo "Warning: No coverage counters file found. Coverage data may not have been collected."; \
		ls -la coverage/; \
	fi

.PHONY: test-coverage
test-coverage: ## Run unit tests with coverage report
	@mkdir -p coverage
	go test ./internal/... -coverprofile=coverage/unit-coverage.out -count=1
	go tool cover -html=coverage/unit-coverage.out -o coverage/unit-coverage.html
	@echo "Unit test coverage report generated: coverage/unit-coverage.html"

.PHONY: test-all
test-all: test test-integration-with-coverage ## Run all tests (unit + integration) with coverage

##@ Docker Compose Workflows

.PHONY: deps-up
deps-up: ## Start dependency services (wiremock) in Docker Compose
	docker compose up -d wiremock

.PHONY: deps-down
deps-down: ## Stop dependency services
	docker compose down

.PHONY: compose-up
compose-up: ## Start all services in Docker Compose (wiremock + app)
	docker compose up --build

.PHONY: compose-down
compose-down: ## Stop all Docker Compose services
	docker compose down

.PHONY: compose-test
compose-test: ## Run outside-in tests in Docker Compose (full integration)
	docker compose -f docker-compose-test.yml up --build --abort-on-container-exit --exit-code-from tests
	docker compose -f docker-compose-test.yml down

##@ Docker Image Management

.PHONY: docker-build-dev
docker-build-dev: ## Build the dev/test Docker image
	docker build --target devtest -t go-outside-in:dev .

.PHONY: docker-build-prod
docker-build-prod: ## Build the production Docker image
	docker build --target prod -t go-outside-in:prod .

.PHONY: docker-build-all
docker-build-all: docker-build-dev docker-build-prod ## Build all Docker images

##@ Code Quality

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint installed)
	@command -v golangci-lint >/dev/null 2>&1 || { echo >&2 "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run

##@ CI-Compatible Workflows

.PHONY: ci-test
ci-test: fmt vet test ## Run all checks and tests (CI-compatible)

.PHONY: ci-test-integration
ci-test-integration: compose-test ## Run integration tests in Docker (CI-compatible)

.PHONY: ci-build
ci-build: docker-build-all ## Build all artifacts (CI-compatible)

.PHONY: ci-full
ci-full: ci-test ci-test-integration ci-build ## Run complete CI pipeline locally

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts and containers
	rm -rf bin/ coverage/ coverage.out coverage.html .server.pid
	docker compose down -v
	docker compose -f docker-compose-test.yml down -v

.PHONY: clean-all
clean-all: clean ## Clean everything including Docker images
	docker rmi go-outside-in:dev go-outside-in:prod 2>/dev/null || true

##@ Common Workflows

.PHONY: local-dev
local-dev: deps-up run-with-mocks ## Start local dev environment (wiremock + local server)

.PHONY: local-test
local-test: deps-up ## Start dependencies and run blackbox tests against local server
	@echo "Starting dependencies..."
	@echo "In another terminal, run: make run-with-mocks"
	@echo "Then run: make test-blackbox-local"

