# Diagrams

This document visualizes the **outside-in testing strategy** across the main runtime contexts (local dev, Docker Compose integration, CI in GitHub Actions, and Kubernetes/prod).

All Mermaid diagrams below use a consistent high-contrast style and a larger font for readability in both light and dark mode.

---

## 1) Testing philosophy and pyramid

```mermaid

graph TD
    subgraph Pyramid["Testing Pyramid (Outside-In Strategy)"]
        E2E["Outside-in / E2E Tests<br/>(HTTP black-box)<br/>PRIMARY CONTRACT"]
        Integration["Integration Tests<br/>(optional, for complex flows)"]
        Unit["Unit Tests<br/>(coverage gaps, complex logic)<br/>SECONDARY"]
    end
    
    E2E -.->|reusable across| Envs["Dev / Compose / CI / K8s"]
    Integration -.->|environment-specific| Limited["Limited use"]
    Unit -.->|internal only| Internal["Internal implementation"]
    
    style E2E fill:#10B981,stroke:#111827,stroke-width:3px,color:#111827
    style Integration fill:#F59E0B,stroke:#111827,stroke-width:2px,color:#111827
    style Unit fill:#EF4444,stroke:#111827,stroke-width:2px,color:#111827
```

- **Outside-in tests are primary**: They validate the stable HTTP contract and are reusable across all environments.
- **Unit tests are secondary**: Used sparingly for coverage gaps and complex internal logic, not for trivial glue code.
- **Integration tests are minimal**: Only when outside-in tests can't adequately cover a scenario.
- **Test portability is critical**: The same test suite runs locally, in CI, and post-deploy with only configuration changes.

## 2) Runtime contexts overview (outside-in first)

```mermaid

flowchart TB
  Client[Client / Test Runner<br/>curl, go test, CI job] -->|HTTP GET /hello| Svc[Go Service<br/>/hello handler]
  Svc -->|HTTP GET| Ext[External HTTP Dependency<br/>real or mocked]
  Svc -->|HTTP 200| Client

  subgraph Contexts["`Runtime Contexts (same interface, different wiring)`"]
    direction LR
    Dev["`Dev Local Binary<br/>(debuggable)`"] --- IT["`Compose Integration<br/>(deps mocked)`"]
    IT --- CI["`GitHub Actions CI<br/>(containerized tests)`"]
    CI --- K8S["`Kubernetes / Prod<br/>(real deps)`"]
  end

  Client -. targets via BASE_URL .-> Contexts
  Ext -. controlled via EXTERNAL_URL .-> Contexts
```

- The only stable contract is the HTTP interface (GET /hello), which is what outside-in tests target.
- Test reuse is enabled by using BASE_URL to point tests at different environments.
- Dependency wiring is controlled via EXTERNAL_URL (e.g., Wiremock in integration; real endpoints in K8s/prod).

## 3) Dockerfile multi-stage build targets

```mermaid

flowchart TB
    Base1[golang:1.24] --> DevTest["devtest target<br/>(go toolchain + delve)"]
    Base2[golang:1.24] --> Build["build target<br/>(static binary)"]
    Build --> BinOut[/out/server binary/]
    Base3[distroless/static] --> Prod["prod target<br/>(distroless)"]
    BinOut -.->|COPY --from=build| Prod
    
    DevTest -.->|used in| ComposeTest[docker-compose-test.yml<br/>tests service]
    Prod -.->|used in| ComposeProd[docker-compose.yml<br/>app service]
    Prod -.->|deployed to| K8s[Kubernetes / Production]
    
    style DevTest fill:#60A5FA,stroke:#111827,stroke-width:2px,color:#111827
    style Build fill:#A78BFA,stroke:#111827,stroke-width:2px,color:#111827
    style Prod fill:#10B981,stroke:#111827,stroke-width:2px,color:#111827
```

- **devtest target**: Contains Go toolchain and optional debugger (dlv), used for running tests in containers.
- **build target**: Intermediate stage that compiles a static binary with trimmed paths.
- **prod target**: Minimal distroless image (~25MB) with only the binary, suitable for production deployment.
- **Separation of concerns**: Test tooling never enters production images, maintaining security and size efficiency.

## 4) Local development (best debugging workflow)

```mermaid

flowchart LR
  Dev[Developer Workstation] -->|docker compose up -d wiremock| WM[Wiremock Container<br/>:8081 -> :8080]
  Dev -->|go run / dlv| SvcLocal["`Go Service (Local Binary)<br/>:8080`"]
  Dev -->|go test ./test/blackbox<br/>BASE_URL=http://localhost:8080| Tests["`Outside-in Tests<br/>(black-box HTTP)`"]

  Tests -->|HTTP GET /hello| SvcLocal
  SvcLocal -->|HTTP GET EXTERNAL_URL| WM
  WM -->|HTTP 204| SvcLocal
  SvcLocal -->|HTTP 200 + message| Tests

```

- Dependencies run in Docker, but the service runs locally, enabling step-through debugging.
- The outside-in tests remain pure HTTP and do not import internal DTOs/models.
- This layout scales directly to “DB + Wiremock” compositions without changing the testing philosophy.

## 5) Integration testing in Docker Compose (service container optional)

```mermaid

flowchart TB
  subgraph Compose[Docker Compose Network]
    WM[Wiremock<br/>wiremock:8080]
    App["`Go Service Container (optional)<br/>app:8080`"]
    TestsC["`Tests Container (devtest target)<br/>BASE_URL=http://app:8080`"]
    TestsC -->|HTTP GET /hello| App
    App -->|HTTP GET EXTERNAL_URL| WM
  end

  Host[Host / CI Runner] -->|docker compose up| Compose
  Host -->|docker compose logs / inspect| Compose

```

- This context is best for reproducible integration (same container graph locally and in CI).
- The "optional service container" pattern lets you choose:
  - local binary for debugging, or
  - containerized service for parity.
- Outside-in tests still treat the service as a black box (HTTP-only).


## 6) GitHub Actions CI pipeline (outside-in gate)

```mermaid

flowchart TD
  GH[GitHub Actions Runner] --> Checkout[Checkout]
  Checkout --> Build["`Build (prod target)<br/>static binary`"]
  Checkout --> Unit[Unit Tests<br/>go test ./...]
  Unit --> Lint["`Lint (optional)<br/>golangci-lint`"]
  Build --> ComposeUp[Compose Up<br/>wiremock + app + tests]
  ComposeUp --> OutsideIn[Outside-in Tests Gate<br/>go test ./test/blackbox]
  OutsideIn --> Report[Publish Results<br/>coverage/artifacts]

```

- CI runs fast inner-loop checks (unit tests, lint) plus an outside-in integration gate.
- The outside-in gate validates the most stable contract: HTTP interface + dependency wiring.
- Because outside-in tests are environment-agnostic, the same suite can run post-deploy (smoke) in K8s.


## 7) Request flow sequence (same test, multiple contexts)

```mermaid

sequenceDiagram
  autonumber
  participant T as Outside-in Test (HTTP client)
  participant S as Go Service (/hello)
  participant E as External Dependency (Wiremock or Real)

  T->>S: GET /hello (User-Agent set)
  S->>E: GET EXTERNAL_URL
  E-->>S: HTTP <status>
  S-->>T: 200 "hello <ua>... got code <status>"

```

- This is the core outside-in contract: tests validate end-to-end behavior without importing internal DTOs/models.
- The external dependency can be swapped via EXTERNAL_URL without changing the test logic.
- This sequence remains stable as the service grows, as long as the public API contract is preserved.


## 8) Kubernetes / production context (post-deploy outside-in smoke)

```mermaid

flowchart LR
  subgraph K8S["Kubernetes Cluster"]
    Ingress["Ingress / Service VIP"] --> Pod["Go Service Pod<br/>/hello"]
    Pod --> Real["Real External Dependency<br/>(prod endpoint)"]
  end

  CI["Post-deploy Job<br/>(GHA)"] -->|"BASE_URL=https://..."| Ingress
  CI -->|"outside-in smoke subset"| Ingress

```

- Post-deploy validation reuses the same outside-in tests, typically a smaller smoke subset.
- Dependencies are real, which catches issues Wiremock cannot (auth, networking, real schemas).
- The testing approach emphasizes contract stability and production wiring correctness.
