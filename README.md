# Outside-In Go Service Example

This repository contains a deliberately small Go HTTP service used to
demonstrate an **outside-in testing strategy** for backend services.

The application exposes a single endpoint:

    GET /hello

The endpoint:
- Reads request metadata (User-Agent)
- Calls an external HTTP dependency
- Responds with a message derived from both inputs

Despite the simplicity of the business logic, the project is structured to
demonstrate a **production-grade testing and delivery strategy** that scales
to larger systems.

---

## What This Project Demonstrates

### 1. Outside-In Testing Strategy

This repository prioritizes **outside-in tests**:
- Tests exercise the service **via its public HTTP interface**
- Dependencies are treated as **black boxes**
- Tests interact with the service the same way real clients do

The testing pyramid is intentionally inverted compared to many traditional
Go codebases:

| Layer | Emphasis |
|-----|---------|
| Outside-in / black-box tests | **High** |
| Integration tests | **High** |
| Unit tests | Moderate / targeted |
| Internal implementation tests | Low |

The goal is to **validate stable interfaces first**, not internal
implementation details.

---

### 2. Test Reuse Across Environments

The same outside-in tests can run against:

- An in-process server (`httptest`)
- A locally running binary (ideal for debugging)
- Docker Compose (mocked dependencies)
- Kubernetes (real dependencies)

This is achieved by:
- Driving tests purely via HTTP
- Using environment variables (e.g. `BASE_URL`) to switch targets
- Avoiding direct imports of application internals in integration tests

---

### 3. Minimal Dependency Surface

The service:
- Uses only the Go standard library
- Avoids HTTP frameworks
- Avoids test frameworks beyond `testing`

This keeps:
- Build times fast
- Container images small
- Behavior transparent

---

### 4. Containerization Strategy

The project uses a **multi-target Dockerfile**:

- **dev/test image**
    - Go toolchain
    - Test and debug tooling
    - Used for CI and local containerized development
- **prod image**
    - Distroless
    - Static binary
    - Non-root execution

This mirrors real-world production constraints while preserving developer
ergonomics.

---

## Outside-In Testing Philosophy

### Why Outside-In?

Outside-in tests:
- Validate the **most stable contracts** (HTTP APIs, schemas, auth behavior)
- Fail when user-visible behavior changes
- Continue to work as internal implementations evolve

As features are added, these tests tend to **break less often** than
implementation-level tests.

---

### Unit Tests Still Matter (But Differently)

This approach does **not** eliminate unit testing.

Instead:
- Unit tests are used **surgically**
- Primarily to:
    - Reach coverage targets
    - Exercise edge cases that are difficult to trigger via HTTP
    - Validate complex logic in isolation

Outside-in tests remain the primary correctness signal.

---

### DTO and Model Handling in Integration Tests

Integration and outside-in tests intentionally:
- **Do not import application DTOs or model classes**
- Parse JSON and validate responses manually

This is intentional.

Benefits:
- Detects drift between models and actual API contracts
- Catches serialization, tagging, and schema mismatches
- Prevents false confidence caused by shared types

---

## Benefits of This Approach

- High confidence in user-visible behavior
- Excellent test reuse across environments
- Strong alignment with production behavior
- Debug-friendly local workflows
- Resistant to internal refactors

---

## Trade-Offs and Risks

- Test setup can become complex
- Slower feedback than pure unit tests
- Requires discipline around environment configuration
- Poorly designed outside-in tests can become overly broad

This approach works best when:
- Service boundaries are well defined
- Configuration is explicit
- Teams value contract stability

---

## When to Use This Pattern

This pattern is especially effective for:
- APIs with external dependencies
- Systems deployed to Kubernetes
- Organizations with strong CI/CD discipline
- Teams that prioritize long-lived, stable tests

---

## Local Development and CI Workflows

### Makefile-Driven Workflows

This repository includes a **self-documenting Makefile** that codifies common
development, testing, and CI workflows.

To see all available commands:

```bash
make help
```

The Makefile provides targets for:
- **Development**: `make build`, `make run`, `make local-dev`
- **Testing**: `make test`, `make test-blackbox-local`, `make test-coverage`
- **Docker Compose**: `make compose-up`, `make compose-test`, `make deps-up`
- **Kubernetes**: `make k8s-create-cluster`, `make k8s-deploy`, `make k8s-full-test`
- **CI-compatible**: `make ci-test`, `make ci-test-integration`, `make ci-full`
- **Code quality**: `make fmt`, `make vet`, `make lint`

---

### CI and Local Parity

The Makefile is designed to **ensure consistency between CI and local
environments**.

Key principles:
- The **same commands** can be run locally and in CI
- No CI-specific scripts or logic
- All workflows are reproducible on a developer's machine

For example:
- `make ci-test` runs formatting, vetting, and all tests
- `make ci-test-integration` runs the full Docker Compose integration suite
- `make ci-full` executes the complete CI pipeline locally

This approach:
- Eliminates "works on my machine" issues
- Makes CI failures easy to reproduce locally
- Keeps CI configuration minimal (just calls `make` targets)

---

### Common Workflows

**Start local development environment:**
```bash
make local-dev
# Starts wiremock in Docker, then runs the server locally
```

**Run tests against a local server:**
```bash
# Terminal 1
make run-with-mocks

# Terminal 2
make test-blackbox-local
```

**Run integration tests in Docker (like CI):**
```bash
make compose-test
```

**Reproduce CI locally:**
```bash
make ci-full
```

**Run tests in a local Kubernetes cluster:**
```bash
# Full automated test (creates cluster, builds, deploys, tests)
make k8s-full-test

# Or step by step:
make k8s-create-cluster    # Create kind cluster
make k8s-build-and-load    # Build and load image
make k8s-deploy            # Deploy to cluster
kubectl port-forward service/go-outside-in 8080:8080 &  # Port forward
make k8s-test              # Run tests
make k8s-delete-cluster    # Clean up
```

---

## CI/CD Workflows

### Pull Request CI (`.github/workflows/ci.yml`)

The CI workflow runs on every pull request and includes:

1. **Lint**: Runs `golangci-lint` with comprehensive checks
2. **Test**: Runs unit tests with formatting and vetting
3. **Integration**: Runs full integration tests in Docker Compose
4. **Build**: Builds both dev and production Docker images
5. **K8s Test**: Deploys to a kind cluster and runs blackbox tests

All jobs run in parallel for fast feedback.

### Deployment Workflow (`.github/workflows/deploy.yml`)

The deployment workflow runs on merge to main and:

1. Builds production Docker image with commit SHA tag
2. **Simulates** pushing to a container registry (echoes commands)
3. **Simulates** deploying to production Kubernetes (echoes kubectl commands)
4. **Simulates** waiting for rollout to complete
5. **Actually** creates a kind cluster and deploys the new version
6. **Actually** runs integration tests against the deployed version
7. Shows what rollback would look like on failure

This workflow demonstrates:
- How to validate deployments with the same tests used in development
- Progressive deployment verification
- Rollback procedures
- The outside-in testing strategy in a production context

---

## Summary

This repository is intentionally small, but the testing and delivery patterns
scale to real systems.

The **outside-in testing strategy** emphasizes:
- Testing what users and clients actually experience
- Reusing tests across environments
- Accepting some setup complexity in exchange for long-term stability

The **Makefile-driven workflows** ensure:
- Consistency between local development and CI
- Self-documenting commands for common tasks
- Easy reproducibility of CI failures

