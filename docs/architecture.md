# Architecture Diagrams

This document visualizes the **outside-in testing strategy** across the main runtime contexts (local dev, Docker Compose integration, CI in GitHub Actions, and Kubernetes/prod).

All Mermaid diagrams below use a consistent high-contrast style and a larger font for readability in both light and dark mode.

---

## 1) Testing philosophy and pyramid

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial'}}}%%
graph TD
    subgraph Test Pyramid
        E2E["E2E Tests<br/>(UI + Backend)<br/>RECOMMENDED / NOT IMPLEMENTED"]@{ shape: tri}
        OutsideIn["Outside-in Integration Tests<br/>(HTTP black-box)<br/>PRIMARY CONTRACT"]
        Unit["Unit Tests<br/>(coverage gaps, complex logic)<br/>SECONDARY"]@{ shape: flip-tri}

        E2E --> OutsideIn --> Unit
    end
    
    E2E -.->|scope| Full["Full Platform, fewer tests"]
    
    OutsideIn -.->|scope| Most["External Contracts, Most tests"]
    Most -.->|reusable across| Envs["Dev / Compose / CI / K8s"]
    Unit -.->|scope| Internal["Internal implementation, fewer tests"]
    
    classDef default fill:#1F2937,stroke:#FFFFFF,stroke-width:2px,color:#FFFFFF
    
    style E2E fill:#374151,stroke:#9CA3AF,stroke-width:2px,stroke-dasharray: 5 5,color:#FFFFFF
    style OutsideIn fill:#10B981,stroke:#000000,stroke-width:4px,color:#000000,font-weight:bold
    style Unit fill:#EF4444,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style Full fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Most fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Envs fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Internal fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
```

- **Outside-in Integration tests are primary**: They validate the stable HTTP contract and are reusable across all environments.
- **Unit tests are secondary**: Used sparingly for coverage gaps and complex internal logic, not for trivial glue code.
- **E2E tests are recommended**: For full system validation (UI, etc.) but are not implemented in this repository.
- **Test portability is critical**: The same test suite runs locally, in CI, and post-deploy with only configuration changes.

## 2) Runtime contexts overview (outside-in first)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial'}}}%%
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

  classDef default fill:#1F2937,stroke:#FFFFFF,stroke-width:2px,color:#FFFFFF
  classDef highlight fill:#E0F2FE,stroke:#000000,stroke-width:3px,color:#000000
  
  class Client,Svc,Ext highlight
  class Dev,IT,CI,K8S default
```

- The only stable contract is the HTTP interface (GET /hello), which is what outside-in tests target.
- Test reuse is enabled by using BASE_URL to point tests at different environments.
- Dependency wiring is controlled via EXTERNAL_URL (e.g., Wiremock in integration; real endpoints in K8s/prod).

## 3) Dockerfile multi-stage build targets

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial'}}}%%
flowchart TB
    Base1[golang:1.24] --> DevTest["devtest target<br/>(go toolchain + delve)"]
    Base2[golang:1.24] --> Build["build target<br/>(static binary)"]
    Base3[golang:1.24] --> BuildCov["build-coverage target<br/>(coverage-instrumented binary)"]
    Build --> BinOut[/out/server binary/]
    BuildCov --> BinCov[/out/server-coverage binary/]
    Base4[distroless/static] --> Prod["prod target<br/>(distroless)"]
    Base5[distroless/static] --> ProdCov["prod-coverage target<br/>(distroless + GOCOVERDIR)"]
    BinOut -.->|COPY --from=build| Prod
    BinCov -.->|COPY --from=build-coverage| ProdCov
    
    DevTest -.->|used in| ComposeTest[docker-compose-test.yml<br/>tests service]
    Prod -.->|used in| ComposeProd[docker-compose.yml<br/>app service]
    ProdCov -.->|used in| ComposeCov[docker-compose-coverage.yml<br/>coverage-instrumented app]
    Prod -.->|deployed to| K8s[Kubernetes / Production]
    
    classDef default fill:#1F2937,stroke:#FFFFFF,stroke-width:2px,color:#FFFFFF
    
    style DevTest fill:#60A5FA,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style Build fill:#A78BFA,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style BuildCov fill:#A78BFA,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style Prod fill:#10B981,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style ProdCov fill:#F59E0B,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style Base1 fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Base2 fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Base3 fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Base4 fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style Base5 fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style BinOut fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style BinCov fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style ComposeTest fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style ComposeProd fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style ComposeCov fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style K8s fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
```

- **devtest target**: Contains Go toolchain and optional debugger (dlv), used for running tests in containers.
- **build target**: Intermediate stage that compiles a static binary with trimmed paths.
- **build-coverage target**: Compiles a binary with `-cover` flag for coverage collection during integration tests.
- **prod target**: Minimal distroless image (~25MB) with only the binary, suitable for production deployment.
- **prod-coverage target**: Similar to prod but runs the coverage-instrumented binary and exposes `/coverage` volume for data collection.
- **Separation of concerns**: Test tooling never enters production images, maintaining security and size efficiency.

## 4) Local development (best debugging workflow)

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial'}}}%%
flowchart LR
  Dev[Developer Workstation] -->|docker compose up -d wiremock| WM[Wiremock Container<br/>:8081 -> :8080]
  Dev -->|go run / dlv| SvcLocal["`Go Service (Local Binary)<br/>:8080`"]
  Dev -->|go test ./test/blackbox<br/>BASE_URL=http://localhost:8080| Tests["`Outside-in Tests<br/>(black-box HTTP)`"]

  Tests -->|HTTP GET /hello| SvcLocal
  SvcLocal -->|HTTP GET EXTERNAL_URL| WM
  WM -->|HTTP 204| SvcLocal
  SvcLocal -->|HTTP 200 + message| Tests

  classDef default fill:#1F2937,stroke:#FFFFFF,stroke-width:2px,color:#FFFFFF
  classDef highlight fill:#FEF3C7,stroke:#000000,stroke-width:3px,color:#000000
  
  class Dev,WM,SvcLocal,Tests highlight
```

- Dependencies run in Docker, but the service runs locally, enabling step-through debugging.
- The outside-in tests remain pure HTTP and do not import internal DTOs/models.

## 5) Coverage Collection Strategy

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial'}}}%%
flowchart TB
    subgraph Unit["`Unit Test Coverage`"]
        UT[go test ./internal/...] -->|generates| UnitCov[coverage/unit-coverage.out]
        UnitCov -->|go tool cover -html| UnitHTML[coverage/unit-coverage.html]
    end

    subgraph Integration["`Integration Test Coverage`"]
        Server[Coverage-Instrumented Server<br/>prod-coverage target] -->|writes to volume| CovDir[/coverage/covcounters.*]
        Tests[Blackbox Tests] -->|HTTP requests| Server
        CovDir -->|go tool covdata textfmt| IntCov[coverage/coverage.out]
        IntCov -->|go tool cover -html| IntHTML[coverage/coverage.html]
    end

    subgraph CI["`CI Workflow`"]
        UnitHTML -.->|artifact| GHA1[GitHub Actions Artifact<br/>unit-test-coverage]
        IntHTML -.->|artifact| GHA2[GitHub Actions Artifact<br/>integration-coverage-report]
    end

    classDef default fill:#1F2937,stroke:#FFFFFF,stroke-width:2px,color:#FFFFFF
    
    style UT fill:#60A5FA,stroke:#000000,stroke-width:2px,color:#000000
    style UnitCov fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style UnitHTML fill:#10B981,stroke:#000000,stroke-width:2px,color:#000000
    style Server fill:#F59E0B,stroke:#000000,stroke-width:3px,color:#000000,font-weight:bold
    style Tests fill:#60A5FA,stroke:#000000,stroke-width:2px,color:#000000
    style CovDir fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style IntCov fill:#E5E7EB,stroke:#000000,stroke-width:2px,color:#000000
    style IntHTML fill:#10B981,stroke:#000000,stroke-width:2px,color:#000000
    style GHA1 fill:#A78BFA,stroke:#000000,stroke-width:2px,color:#000000
    style GHA2 fill:#A78BFA,stroke:#000000,stroke-width:2px,color:#000000
```

- **Unit coverage** (`make test-coverage`): Traditional Go test coverage for internal packages.
- **Integration coverage** (`make test-integration-with-coverage`): Uses a coverage-instrumented binary built with `-cover` flag, deployed in Docker Compose with dependencies.
- **CI artifacts**: Both coverage reports are uploaded to GitHub Actions for download and review.
- **Separation**: Unit tests cover internal logic; integration tests measure coverage from HTTP interface exercising real request flows.
- **Volume mounting**: The `prod-coverage` image writes coverage data to `/coverage` which is mounted from the host for post-test analysis.

