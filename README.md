# BookNest Platform

Go backend for the BookNest bookstore platform. The service uses Gin for HTTP routing, PostgreSQL for persistence, GORM plus `pgx` for data access, Redis for rate limiting support, and Swagger for local API exploration.

## What this service provides

- Versioned REST API mounted at `/api/v1`
- JWT-based authentication with access-token refresh support
- Catalog management for books, authors, categories, and publishers
- Cart and checkout flows
- User and admin order views
- Swagger UI protected by basic auth

## Project structure

- `main.go`: application startup, env loading, Swagger registration
- `main_setup.go`: dependency wiring, middleware, route mounting, HTTP server
- `internal/http/api/v1`: active API version
- `internal/http/controller`: HTTP handlers and route registration
- `internal/service`: business logic
- `internal/repository`: PostgreSQL/GORM data access
- `internal/http/database/migrations`: SQL migrations
- `internal/middleware`: auth, rate limiting, logging, error handling
- `docs`: generated Swagger assets

## Prerequisites

- Go `1.24+`
- PostgreSQL `15+`
- Redis `7+` for local parity with the current rate-limiter setup
- `migrate` CLI for running SQL migrations manually
- `swag` CLI if you need to regenerate Swagger docs

## Environment

Create a local `.env` file in this directory. Use your own secrets; do not commit real credentials.

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=booknest
DB_NAME=booknest
DB_PORT=5432
DB_URL=postgres://postgres:booknest@localhost:5432/booknest?sslmode=disable

JWT_SECRET_V1=replace-me
JWT_REFRESH_V1=replace-me

# Optional backward compatibility values if an older client still expects v0 keys
JWT_SECRET_V0=
JWT_REFRESH_V0=

SWAGGER_USER=booknest
SWAGGER_PASSWORD=replace-me

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Optional: overrides the host shown in Swagger docs
API_HOST=localhost:8080
```

Notes:

- `godotenv.Load()` is required during startup, so `.env` needs to exist locally.
- If `REDIS_ADDR` is empty, the app falls back to an in-memory login rate limiter.
- Local CORS currently allows `http://localhost:3000` and `http://localhost:5173`.

## Local setup

1. Start PostgreSQL and create the `booknest` database.
2. Start Redis on `localhost:6379`.
3. Add the `.env` file shown above.
4. Install dependencies and run migrations.

```bash
go mod download
make migrate-up
```

## Run the API

```bash
go run .
```

The service starts on `http://localhost:8080`.

Useful endpoints:

- Health check: `GET /health`
- API base: `http://localhost:8080/api/v1`
- Swagger UI: `http://localhost:8080/swagger/v1/index.html`

## Common commands

```bash
go test ./...
make migrate-up
make migrate-down
make migrate-version
make migrate-force VERSION=<n>
make migrate-new name=<migration_name>
make doc
```

## API surface

The current frontend expects these routes under `/api/v1`:

- Auth: `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`, `/auth/reset-password/confirm`
- Books: `/books`, `/book/:id`
- Cart: `/cart`, `/cart/items`, `/cart/items/:book_id`, `/cart/clear`
- Orders: `/orders`, `/orders/checkout`, `/orders/confirm`, `/admin/orders`
- Catalog admin: `/authors`, `/categories`, `/publishers`

There is also a legacy `/refresh` fallback handled by the frontend for older environments, but the current backend route is `/auth/refresh`.

## Swagger

- Swagger is registered as the named instance `booknest-v1`
- The documented base path is `/api/v1`
- Access is protected by `SWAGGER_USER` and `SWAGGER_PASSWORD`
- Regenerate docs after annotation changes with `make doc`

## Docker

The repository includes a `docker-compose.yml` for local infrastructure:

```bash
docker compose up --build
```

This starts:

- the API container
- PostgreSQL 15
- Redis 7

If you rely on Docker for first-time setup, make sure your `.env` matches the containerized database credentials before starting the stack.
