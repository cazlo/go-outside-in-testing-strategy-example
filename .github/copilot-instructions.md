# Copilot / Agent Instructions

This repository demonstrates an **outside-in testing strategy** for Go
services.

Agentic AI operating in this repository must respect the architectural and
testing constraints described below.

---

## Core Architectural Intent

- The primary interface is HTTP
- External dependencies are accessed via HTTP
- Tests prioritize validating behavior at the interface boundary
- Internal implementation details are considered volatile

Avoid optimizing for internal testability at the expense of interface clarity.

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

## When Adding New Features

Agents should:
1. Add or extend outside-in tests first
2. Validate behavior via HTTP
3. Add unit tests only if needed for coverage or complex logic
4. Preserve test reuse across environments

Do not introduce:
- Environment-specific test logic
- Hidden coupling between tests and internal implementation

---

## Summary for Agents

This repository intentionally favors:
- Contract testing over implementation testing
- Stability over test granularity
- Reuse over simplicity

All changes should reinforce the **outside-in testing strategy** and avoid
regressions in test portability, clarity, or runtime parity.
