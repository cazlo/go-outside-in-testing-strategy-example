#SHELL := /bin/bash
#.ONESHELL:
#.SHELLFLAGS := -eu -o pipefail -c
#
#ASDF_DIR ?= $(HOME)/.asdf
#
## Ensure asdf is initialized and shims are on PATH for all recipes
#export PATH := $(ASDF_DIR)/bin:$(ASDF_DIR)/shims:$(PATH)
#
## If you want the asdf function available too (for `asdf ...` commands)
#define ASDF_INIT
#. "$(ASDF_DIR)/asdf.sh"
#endef

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

GOLANGCI_BIN := ./bin/golangci-lint

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
	go test ./internal/... -coverpkg=./internal/... -count=1 -v -cover

.PHONY: test-unit
test-unit: test ## Alias for test (unit tests only)

.PHONY: test-blackbox
test-blackbox: ## Run blackbox tests against a running server (requires BASE_URL)
	go test ./test/blackbox -count=1 -v

.PHONY: test-blackbox-local
test-blackbox-local: ## Run blackbox tests against local server (assumes server running on :8080)
	BASE_URL=http://localhost:8080 WIREMOCK_URL=http://localhost:8081 go test ./test/blackbox -count=1 -v

.PHONY: test-integration
test-integration: deps-up ## Run integration tests with dependencies (server must be running separately)
	@echo "Starting wiremock dependency..."
	@echo "Run 'make run-with-mocks' in another terminal, then 'make test-blackbox-local'"
	@echo "Or use 'make test-integration-with-coverage' for automated testing with coverage"

.PHONY: build-coverage
build-coverage: ## Build server binary with coverage instrumentation
	go build -cover -o bin/server-coverage ./cmd/server

# todo below kinda works but still needs help
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
	go test ./internal/... -coverpkg=./internal/... -coverprofile=coverage/unit-coverage.out -count=1
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

.PHONY: install-golangci-lint
install-golangci-lint: ## Install golangci-lint
	$(ASDF_INIT)
	@echo "Installing golangci-lint..."
	@echo $(go env GOPATH)
	# see https://golangci-lint.run/docs/welcome/install/local/
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.7.2

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: ensure-golangci-lint
ensure-golangci-lint:
	@if [ ! -x $(GOLANGCI_BIN) ]; then \
		echo "golangci-lint not found, installing..."; \
		$(MAKE) install-golangci-lint; \
	fi

.PHONY: lint
lint: ensure-golangci-lint ## Run golangci-lint (will install if not found)
	$(GOLANGCI_BIN) run

.PHONY: lint-fix
lint-fix: ensure-golangci-lint ## Run golangci-lint with autofix mode
	$(GOLANGCI_BIN) run --fix

##@ CI-Compatible Workflows

.PHONY: ci-test
ci-test: fmt vet test ## Run all checks and tests (CI-compatible)

.PHONY: ci-test-integration
ci-test-integration: compose-test ## Run integration tests in Docker (CI-compatible)

.PHONY: ci-build
ci-build: docker-build-all ## Build all artifacts (CI-compatible)

.PHONY: ci-full
ci-full: ci-test ci-test-integration ci-build ## Run complete CI pipeline locally

##@ Kubernetes Workflows

.PHONY: k8s-create-cluster
k8s-create-cluster: ## Create a local kind cluster for testing
	@echo "Creating kind cluster..."
	@kind create cluster --name test-cluster --wait 30s || echo "Cluster may already exist"

.PHONY: k8s-delete-cluster
k8s-delete-cluster: ## Delete the local kind cluster
	@echo "Deleting kind cluster..."
	@kind delete cluster --name test-cluster

.PHONY: k8s-build-and-load
k8s-build-and-load: ## Build Docker image and load into kind cluster
	@echo "Building Docker image..."
	docker build --target prod -t go-outside-in:test .
	@echo "Loading image into kind cluster..."
	kind load docker-image go-outside-in:test --name test-cluster

.PHONY: k8s-deploy
k8s-deploy: ## Deploy application and dependencies to kind cluster
	@echo "Cluster Context"
	kubectl cluster-info
	kubectl config current-context
	@echo "Deploying wiremock..."
	kubectl apply -f k8s/wiremock.yaml
	@echo "Waiting for wiremock to be ready..."
	kubectl wait --for=condition=ready pod -l app=wiremock --timeout=60s
	@echo "Deploying application..."
	kubectl apply -f k8s/app.yaml
	@echo "Waiting for application rollout..."
	kubectl rollout status deployment/go-outside-in --timeout=90s
	kubectl wait --for=condition=ready pod -l app=go-outside-in --timeout=60s
	@echo "Deployment complete!"

.PHONY: k8s-test
k8s-test: ## Run blackbox tests against kind cluster (requires port-forward in another terminal)
	@echo "Testing against kind cluster..."
	@echo "Make sure to run 'kubectl port-forward service/go-outside-in 8080:8080' and 'kubectl port-forward service/wiremock 8081:8080' in another terminal"
	BASE_URL=http://localhost:8080 WIREMOCK_URL=http://localhost:8081 go test ./test/blackbox -count=1 -v

.PHONY: k8s-test-ci
k8s-test-ci: ## Run blackbox tests against kind cluster (CI-compatible with auto port-forward)
	@echo "Starting port-forward in background..."
	@kubectl port-forward service/go-outside-in 8080:8080 > /dev/null 2>&1 & echo $$! > .k8s-port-forward-app.pid
	@kubectl port-forward service/wiremock 8081:8080 > /dev/null 2>&1 & echo $$! > .k8s-port-forward-wiremock.pid
	@sleep 5
	@echo "Running blackbox tests..."
	@BASE_URL=http://localhost:8080 go test ./test/blackbox -count=1 -v || (kill `cat .k8s-port-forward-app.pid` 2>/dev/null; rm -f .k8s-port-forward-app.pid; exit 1)
	@kill `cat .k8s-port-forward-app.pid` 2>/dev/null || true
	@kill `cat .k8s-port-forward-wiremock.pid` 2>/dev/null || true
	@rm -f .k8s-port-forward-app.pid
	@rm -f .k8s-port-forward-wiremock.pid
	@echo "✓ Kubernetes tests passed!"

.PHONY: k8s-full-test
k8s-full-test: k8s-create-cluster k8s-build-and-load k8s-deploy k8s-test-ci ## Full K8s test: create cluster, build, deploy, test (CI-compatible)

.PHONY: k8s-test-with-cluster
k8s-test-with-cluster: k8s-full-test ## Alias for k8s-full-test

.PHONY: k8s-logs
k8s-logs: ## Show logs from all pods
	@echo "=== Application logs ==="
	@kubectl logs -l app=go-outside-in --tail=50
	@echo ""
	@echo "=== Wiremock logs ==="
	@kubectl logs -l app=wiremock --tail=50

.PHONY: k8s-debug
k8s-debug: ## Show debugging information for Kubernetes deployment
	@echo "=== Pods ==="
	@kubectl get pods
	@echo ""
	@echo "=== Services ==="
	@kubectl get services
	@echo ""
	@echo "=== Deployments ==="
	@kubectl get deployments
	@echo ""
	@echo "=== Pod descriptions ==="
	@kubectl describe pods

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

