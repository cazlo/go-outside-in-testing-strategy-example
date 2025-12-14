# Extending to Database-Backed Services

The outside-in testing pattern demonstrated in this repository (using HTTP dependencies) extends naturally to services that use a database.

## The Pattern

Instead of mocking the database driver or using an in-memory database (like SQLite) for tests, the outside-in approach encourages running a **real database instance** (e.g., Postgres, MySQL) in a container, just like any other external dependency.

### Key Differences

1.  **Dependency Management**:
    *   **HTTP Service**: You might use Wiremock to stub external API calls.
    *   **Database**: You spin up a real database container (e.g., `postgres:15-alpine`) in Docker Compose or your test environment.

2.  **State Management**:
    *   **HTTP Service**: Wiremock state is usually reset via API calls or is stateless per request.
    *   **Database**: You need a strategy to manage data state between tests.
        *   **Transaction Rollback**: Start a transaction before each test and roll it back after. This is fast but can be complex to implement if the application manages its own transactions.
        *   **Truncate/Clean**: Truncate all tables between tests. This is cleaner but slower.
        *   **Unique Data**: Generate unique IDs for every test case so they don't collide. This allows running tests in parallel but requires discipline.

3.  **Configuration**:
    *   Just like `EXTERNAL_URL` configures the HTTP client, a `DATABASE_URL` environment variable configures the database connection.
    *   Tests use this same `DATABASE_URL` to seed data or verify side effects directly in the DB if necessary (though verifying via the API is preferred).

## Why Real Databases?

*   **Fidelity**: In-memory mocks often behave differently than real databases (e.g., locking, constraints, specific SQL syntax).
*   **Confidence**: You are testing the actual SQL queries and driver interactions that will run in production.
*   **Simplicity**: You don't need to maintain complex mocks of database interfaces.

## Example Workflow

1.  **`make deps-up`**: Starts Postgres in Docker.
2.  **`make run`**: Starts the service locally, connected to the Dockerized Postgres.
3.  **`make test-blackbox-local`**: Runs tests.
    *   Test A: POST /users (creates user) -> 201 Created.
    *   Test A: GET /users/{id} -> 200 OK (verifies persistence).
