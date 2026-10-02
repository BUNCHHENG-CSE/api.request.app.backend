# FlowAPI Backend

A Go/Gin and PostgreSQL backend for storing API workspaces, flows, request definitions, environments, and collections.

## Current capabilities

- User registration and JWT login
- Bcrypt password hashing
- Authenticated current-password verification and password changes
- Authenticated workspace CRUD
- CRUD for flows, stored requests, environments, and collections
- Backend execution of stored HTTP requests
- Environment-variable substitution and persistent execution history
- Request duplication, bulk deletion, and sequential flow runs
- Workspace-scoped global search
- SSRF-aware outbound networking, CORS, security headers, and graceful shutdown
- Hybrid PostgreSQL plus local Bruno-style `.bru` collection storage
- Collection/request association and flow graph fields
- Owner checks across all resource operations
- PostgreSQL persistence through GORM
- Docker Compose development setup
- Health check at `GET /health`

This is an early API, not yet a complete Postman-style collaboration and automation platform. See [API_DOCUMENTATION.md](API_DOCUMENTATION.md) for the exact endpoint contract and remaining gaps.

## Run locally

Create a `.env` file containing at least:

```dotenv
SERVER_PORT=8080
JWT_SECRET=replace-with-a-long-random-secret
DATABASE_URL=host=localhost user=postgres password=secret dbname=request_app port=5432 sslmode=disable
ALLOWED_ORIGINS=http://localhost:3000
ALLOW_PRIVATE_REQUESTS=false
LOCAL_COLLECTIONS_ROOT=./data/collections
```

Then start the database and API:

```bash
docker compose up --build
```

For a Go process running outside Docker, make sure `DATABASE_URL` uses `host=localhost`. Inside Docker Compose, use `host=db`.

## Development checks

```bash
go test ./...
go vet ./...
```

## Project layout

```text
cmd/api/                         application entry point
internal/domain/                 entities and repository contracts
internal/application/            business services
internal/infrastructure/         PostgreSQL repositories and configuration
internal/interfaces/api/         Gin handlers, middleware, responses, and routes
```

## Before production

- Add pagination, filtering, and ordering as required by the client.
- Configure an outbound domain allowlist if deployments require stricter egress policy.
- Replace automatic schema migration with versioned migrations.
- Add CORS policy, rate limiting, structured logs, request IDs, and graceful shutdown.
- Store environment secrets encrypted and avoid returning them from list endpoints.
