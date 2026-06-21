# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`booknest-backend` is a Go 1.24 REST API for the BookNest online bookstore. It uses Gin, PostgreSQL with pgvector, GORM + pgx, Redis (for login rate limiting), AWS SES/SNS/S3, an OpenAI-compatible AI provider, and golang-migrate.

## Build, run, test

All commands assume you have a `.env` at the repo root (see README for the full key list — at minimum `JWT_SECRET_V1` and one SES or SNS notification provider must be set).

```bash
go mod download
make migrate-up        # runs internal/http/database/migrations
make run               # go run . — serves on :8080
go test ./...          # all tests
go test ./internal/middleware/...   # one package
go test -run TestX -count=1 ./internal/path/...   # single test
go test ./internal/http/controller -run TestAI   # filter by name
make doc               # regenerate Swagger (swag init --parseDependency --parseInternal)
```

testing refer: @TESTING.md file

The Dockerfile runs `migrate up` against `DB_URL` via `entrypoint.sh` before starting the binary, so `make migrate-up` only needs to be run manually for local `go run`.

## Architecture

### Wiring (`main.go`, `main_setup.go`)

- `main.go` loads `.env`, calls `database.Connect()` (pgxpool), then `SetupServer` and `StartHTTPServer`.
- `main_setup.go` is the dependency graph. `SetupServer(dbpool)` builds GORM, JWT config, Redis, every repo, every service, every controller, and assembles the Gin engine. The order matters: AI provider is constructed best-effort; `bookService` depends on `aiService`; order service can be wrapped by `remoteOrderService` (gRPC) when `ORDER_SERVICE_MODE=microservice`.

### Layers (each top-level `internal/<layer>` package has one subdir per resource)

- `internal/domain` — interfaces and DTOs. Every service, repo, and controller implements a `domain.<X>Controller` / `domain.<X>Service` / `domain.<X>Repository` interface declared here. **The domain package is the contract; new code adds an interface there before implementing it.**
- `internal/repository` — pgxpool + GORM data access. The `DBExecer` interface in `internal/domain/transaction.go` lets services run inside a `TransactionManager.InTransaction` callback without importing pgx. `util.NewTransactionManager(pool)` is the only tx source.
- `internal/service` — business logic. Constructed via `NewXService(...)` factories in `main_setup.go`.
- `internal/http/controller` — Gin handlers plus `*_routes.go` files with `RegisterXRoutes(r, jwtConfig, controller)`. Per-resource wiring lives in `register_routes_test.go` and route constants live in `internal/http/routes/routes.go`.
- `internal/http/api` — version routing. `MountVersions(r, registrars...)` mounts each `VersionRegistrar` under `/api/{version}`. Only `v1` is live; `v2` is an empty scaffold. `v1.Router` holds the controller interfaces, **not** the concrete types — keep that boundary.
- `internal/middleware` — JWT, roles, rate limiting, logging, CORS, error handler, security headers, Swagger basic auth.
- `internal/ai/provider` (preferred) and `internal/pkg/ai` (legacy mirror) — OpenAI provider. `internal/ai/provider` uses `openai-go` and supports `Generate` + `Embed`; `internal/pkg/ai` is a raw-HTTP `Generate`-only clone kept for the AI health endpoint. New code should call into `internal/ai/provider` via `domain.AIService`.
- `internal/pkg/storage` — S3 image upload helper used by `imageController`.
- `internal/pkg/util` — `TransactionManager`, misc helpers.
- `third_party/booknest-order-service` — vendored Go module imported as `booknest-order-service/gen/order/v1`. The `replace` directive in `go.mod` points here so the build never reaches the network for proto types.

### Versioned API surface

All routes mount under `/api/v1` (no legacy unversioned routes — `main_test.go` asserts `/login` returns 404). JWT is verified by the named-key `kid` in the token header against `JWT_SECRET_V1` / `JWT_SECRET_V0`; `JWTAuthMiddleware` injects `user_id`, `email`, `user_role` into the gin context. `RequireAdmin()` reads `user_role` and gates admin-only routes (admin book CRUD, admin order status, image upload).

### Order service has two backends

`order_service.NewOrderServiceWithNotification` is the monolith. When `ORDER_SERVICE_MODE=microservice` (or `USE_ORDER_MICROSERVICE=true`) and `ORDER_SERVICE_GRPC_ADDR` is set, `SetupServer` wraps it with `NewRemoteOrderService` which forwards the gRPC methods to `BookNest-OrderService`. The gRPC client interface (`grpcOrderClient`) is defined in `remote_order_service.go` to keep tests hermetic — tests stub the interface, not the conn.

### AI / semantic features

- `AIService.Chat` does intent classification via the OpenAI provider, then dispatches: `get_book`, `get_books_by_category`, `chat`, `semantic_search`, `recommendation`. Search and recommendations go through `book_embedding_repo` + `SearchNearestBooks` (pgvector `<->` operator).
- `domain.EmbeddingVector` is a `[]float64` with a custom `driver.Valuer`/`Scanner` that serialises to the pgvector text format. Migration `20260601120000_add_book_embeddings` is the source of truth for the `vector(1536)` column.
- `internal/ai/embedding_text.go` formats the input to the embedding model: `Title\n\nDescription\n\nSummary\n\nCategory1 Category2 ...`. Use it from anywhere you need a deterministic embedding input — do not reinvent the format.
- `book_embedding_service` regenerates embeddings on book create/update via the AI provider. It is wired in `main_setup.go` after `aiService` and before `bookService`.

### Swagger

- Registered as the named instance `booknest-v1` (see `main.go` `configureSwagger`). `BasePath` is `/api/v1`; access is gated by `SwaggerAuthMiddleware` (basic auth from `SWAGGER_USER`/`SWAGGER_PASSWORD`).
- Comments above handlers and structs drive generation. Run `make doc` after editing annotations; do not hand-edit files in `docs/`.

## Migrations

- Files live in `internal/http/database/migrations`, named `<utc-timestamp>_<name>.{up,down}.sql`. Use `make migrate-new name=<name>` (the Makefile uses `migrate create`).
- The `MIGRATIONS_DIR` and `DB_URL` come from the included `.env` (`make` `-include .env`).
- Always add both `.up.sql` and `.down.sql`. New columns need a down path that drops them; new tables need both directions.

## Conventions worth knowing

- `internal/http/controller` exposes a `SetJWTConfig` / `SetRedisClient` / `getJWTConfig` / `getRedisClient` pair because Gin route registration functions take a `JWTConfig` argument directly. Tests use these to inject a config without touching env.
- Tests that need a real GORM handle swap the package-level `connectGORM` in `main_setup.go` for an in-memory SQLite or pgxmock; use that pattern instead of mocking out individual repos when you want full coverage of `SetupServer` (`main_test.go` shows both success and failure paths).
- `domain.BookEmbeddingRepository` is the interface used by both `aiService` and `bookService` — do not pass concrete repo types between them.
- `internal/ai/provider` and `internal/pkg/ai` overlap by design (provider has embeddings; pkg only has chat). When touching AI code, prefer `internal/ai/provider` and add the new method there.
- Notifications: `initNotificationService` in `main_setup.go` returns an error if neither SES nor SNS is configured — there is no silent fallback. If a test or local mode needs to bypass this, inject via `NewNotificationServiceWithProvidersAndRepository` instead of changing env.
- `git status` shows in-flight work on AI controller tests (`internal/http/controller/ai_controller_test.go` is currently failing on `main` — 401 where 200/400/503 is expected; the `controller.service == nil` guard runs after the JWT middleware in the test setup). Don't assume `go test ./...` is green until that branch is reconciled.

## Things to double-check before changing

- Migrations: do not edit a migration that has already shipped; add a new one.
- JWT keys: `JWT_SECRET_V0` and `JWT_SECRET_V1` are looked up by the token's `kid` header claim. Rotation = bump `_V1` and move the old secret to `_V0`; both must be set or tokens from the previous key fail.
- `ORDER_SERVICE_MODE=microservice` requires `ORDER_SERVICE_GRPC_ADDR`; `SetupServer` returns an error otherwise.
- AI features: `OPENAI_API_KEY` is required at startup for `aiService` to do anything; the provider itself is constructed best-effort (warn, don't fail) so the server still boots without it.
</content>
</invoke>