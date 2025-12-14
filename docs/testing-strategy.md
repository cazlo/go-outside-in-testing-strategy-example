# Outside-In Testing Strategy

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

## Test Reuse Across Environments

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
