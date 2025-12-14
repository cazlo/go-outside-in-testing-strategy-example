# Outside-In Go Service Example

This repository demonstrates an **outside-in testing strategy** for Go backend services. It prioritizes testing the service via its public HTTP interface, treating dependencies as black boxes, and ensuring tests are reusable across different environments (Local, Docker, Kubernetes).

## Documentation

Detailed documentation has been moved to the `docs/` directory:

- [Architecture & Diagrams](docs/architecture.md)
- [Testing Strategy & Philosophy](docs/testing-strategy.md)
- [Database Patterns](docs/database-pattern.md)
- [Deployment & Kubernetes](docs/deployment.md)

## Getting Started

This project uses a `Makefile` to handle all build, run, and test commands.

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

### Common Commands

To see all available commands, run:
```bash
make help
```

#### Development
```bash
# Start dependencies (Wiremock)
make deps-up

# Run the service locally
make run

# Run the service locally with mocked dependencies
make run-with-mocks
```

#### Testing
```bash
# Run unit tests
make test

# Run black-box integration tests against the local service
make test-blackbox-local

# Run all tests
make test-all
```

#### Build & Clean
```bash
# Build the binary
make build

# Clean up artifacts
make clean
```

