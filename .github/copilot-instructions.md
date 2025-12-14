# Copilot / Agent Instructions

This repository demonstrates an **outside-in testing strategy** for Go
services.

Agentic AI operating in this repository must respect the architectural and
testing constraints described below.

---

## ⚠️ CRITICAL: Always Use the Makefile

**DO NOT run `go test`, `go build`, or `go run` commands directly.**

**ALWAYS use Makefile targets** (e.g., `make test`, `make build`, `make run-with-mocks`).

The Makefile handles environment setup, dependency orchestration, coverage
instrumentation, and process lifecycle management. Bypassing it will cause
failures.

See the "Makefile Usage" section below for available targets.

---

## Core Architectural Intent

- The primary interface is HTTP
- External dependencies are accessed via HTTP
- Tests prioritize validating behavior at the interface boundary
- Internal implementation details are considered volatile

Avoid optimizing for internal testability at the expense of interface clarity.

---

## Documentation & Patterns

Detailed architectural patterns and strategies are documented in the `docs/` directory.
Consult these files for deep dives into specific topics:
- `docs/architecture.md`: Visual diagrams of the testing pyramid and runtime contexts.
- `docs/testing-strategy.md`: Detailed philosophy on outside-in testing.
- `docs/database-pattern.md`: How to handle database dependencies (real containers vs mocks).
- `docs/deployment.md`: Deployment strategies for PRs and production.

---

## Outside-In Testing Strategy (Required Context)

### Key Principles

1. **Outside-in tests are primary**
    - Prefer black-box HTTP tests over internal unit tests
    - Tests should behave like real clients

2. **Test reuse is intentional**
    - The same tests must be able to run:
        - locally
        - in Docker Compose
        - in Kubernetes
    - Configuration, not code branching, controls environment differences

3. **Unit tests are secondary**
    - Use unit tests sparingly
    - Focus on coverage gaps and complex internal logic
    - Do not over-test trivial glue code

---

## Testing Rules for Agents

When adding or modifying tests:

- Do **not** import application DTOs or model structs into
  integration or outside-in tests
- Parse JSON responses directly and validate fields explicitly
- Treat the service as a black box whenever possible
- Prefer HTTP-level assertions over function-level assertions

This is intentional to catch:
- serialization issues
- schema drift
- contract mismatches

---

## Service Design Constraints

- Prefer standard library packages
- Avoid introducing frameworks unless explicitly justified
- Keep handlers thin
- Push complexity into testable components

External calls:
- Must be configurable via environment variables
- Must be replaceable with mock endpoints (e.g. Wiremock)

---

## Database Strategy

When adding database interactions:
- **Do NOT use in-memory mocks** (like sqlite or go-sqlmock) for outside-in tests.
- **Use real database containers** (e.g., Postgres in Docker) managed via `make deps-up`.
- Tests must manage their own state (e.g., unique IDs or cleanup) to allow parallel execution where possible.
- Configuration must be via `DATABASE_URL` env vars.

---

## Containerization Constraints

- Production images must remain distroless
- No test tooling in production images
- Debugging support belongs in dev/test images only

Do not:
- Add shell utilities to prod images
- Assume root access
- Add runtime dependencies without strong justification

---

## Local Development Expectations

Local workflows must support:
- Running dependencies in Docker
- Running the service locally with a debugger
- Running outside-in tests against a locally running service

Avoid designs that require:
- Full container rebuilds to debug logic
- Test-only code paths in production binaries

---

## Makefile Usage (CRITICAL)

**ALWAYS use the Makefile for running tests and building the project.**

Do NOT run `go test` commands directly. Instead, use the appropriate Make targets:

### Testing Commands
- **Unit tests**: `make test` or `make test-unit`
- **Blackbox tests (local)**: `make test-blackbox-local`
- **Integration tests with coverage**: `make test-integration-with-coverage`
- **All tests**: `make test-all`
- **CI test suite**: `make ci-test`
- **Docker Compose tests**: `make compose-test`
- **Kubernetes tests**: `make k8s-full-test`

### Development Commands
- **Build**: `make build`
- **Run locally**: `make run`
- **Run with mocks**: `make run-with-mocks`
- **Start dependencies**: `make deps-up`
- **Stop dependencies**: `make deps-down`

### Other Useful Targets
- **Format code**: `make fmt`
- **Lint**: `make lint`
- **Lint autofix**: `make lint-fix`
- **Clean**: `make clean`
- **Help**: `make help`

The Makefile handles:
- Environment setup
- Dependency orchestration
- Coverage instrumentation
- Process management
- Cleanup

**Never bypass the Makefile** - it contains critical setup and teardown logic.

---

## Linting Expectations

- Run `make lint` before sharing code and `make lint-fix` whenever `golangci-lint` can correct style issues automatically.
- Linting is strict about `errcheck`: always inspect or propagate errors. If dismissal is intentional, log the failure explicitly to satisfy the rule.
- Avoid "blank" helper functions that swallow errors; even housekeeping work (like closing resources) must log failures so issues surface during debugging.
- Prefer small helper functions for repeated cleanup logic instead of anonymous defers that hide unchecked errors.

---

## When Adding New Features

Agents should:
1. Add or extend outside-in tests first
2. Validate behavior via HTTP
3. Add unit tests only if needed for coverage or complex logic
4. Preserve test reuse across environments
5. **Use Makefile targets for all test execution**

Do not introduce:
- Environment-specific test logic
- Hidden coupling between tests and internal implementation
- Direct `go test` invocations (use Makefile instead)

---

## Summary for Agents

This repository intentionally favors:
- Contract testing over implementation testing
- Stability over test granularity
- Reuse over simplicity

All changes should reinforce the **outside-in testing strategy** and avoid
regressions in test portability, clarity, or runtime parity.

### Critical Rules

1. **NEVER run `go test` directly** - always use `make test`, `make test-blackbox-local`, etc.
2. **NEVER bypass the Makefile** for building, running, or testing
3. Tests must validate HTTP behavior, not internal implementation
4. Parse JSON responses directly, don't import application types into tests
5. Environment variables, not code, control configuration differences
6. **Use real database containers** for tests, not in-memory mocks

