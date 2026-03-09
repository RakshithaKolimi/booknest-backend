# BookNest Platform

Backend API for BookNest (Go + Gin + PostgreSQL + GORM/pgx).

For the repo-wide production roadmap and interview-focused upgrade plan, see `../PRODUCTION_NEXT_STEPS.md`.

## Architecture

- `main.go`, `main_setup.go`: bootstrap + dependency wiring
- `internal/http/api`: API version mounting contract (`/api/{version}`)
- `internal/http/api/v1`: full v1 route registration (current production version)
- `internal/http/api/v2`: scaffold only (ready for future endpoints)
- `internal/http/controller`: HTTP handlers and route registration
- `internal/service`: business logic
- `internal/repository`: DB access
- `internal/domain`: entities, enums, interfaces
- `internal/http/database/migrations`: SQL migrations
- `internal/middleware`: JWT auth, admin role guard, error/logging middleware

## Prerequisites

- Go 1.24+
- PostgreSQL 15+
- `migrate` CLI (for local DB migrations)

## Environment

Create/update `.env` in this folder with:

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=booknest
DB_NAME=booknest
DB_PORT=5432
DB_URL=postgres://postgres:booknest@localhost:5432/booknest?sslmode=disable
JWT_SECRET=booknest_secret
SWAGGER_USER=booknest
SWAGGER_PASSWORD=<your-password>
JWT_SECRET_V1=
JWT_REFRESH_V1=
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

Note: `JWT_SECRET_V0` and `JWT_REFRESH_V0` is still supported for backward compatibility, but `JWT_SECRET_V1` and `JWT_REFRESH_V1`  are the primary keys.

## Run (Interview-Safe)

From this folder:

```bash
go mod download
go test ./...
go run .
```

- API: `http://localhost:8080`
- Health: `GET /health`
- v1 API base: `/api/v1`
- Swagger (v1): `http://localhost:8080/swagger/v1/index.html`

## API Versioning

- Only `v1` is implemented and mounted in runtime.
- `v2` folder exists as a plug-in scaffold but is intentionally not mounted.
- To introduce `v2`, implement handlers/services and mount the v2 registrar in `main_setup.go`.

## Swagger (Production Notes)

- Swagger is served as a named instance (`booknest-v1`) for version isolation.
- Runtime base path is set to `/api/v1`.
- Host can be overridden using `API_HOST` (useful behind reverse proxies/gateways).
- Swagger endpoint is protected by basic auth (`SWAGGER_USER` / `SWAGGER_PASSWORD`).

## Migrations

Using Makefile targets:

```bash
make migrate-up
make migrate-down
make migrate-version
```

## Docker Option

```bash
docker compose up --build
```

Starts API + Postgres and runs migrations via `entrypoint.sh`.
