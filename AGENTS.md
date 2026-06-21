# Repository Guidelines

## Project Structure & Module Organization

BookNest Backend is a Go 1.24 REST API. Startup and dependency wiring live in `main.go` and `main_setup.go`. Domain models are in `internal/domain`, HTTP routing and handlers in `internal/http/api` and `internal/http/controller`, business logic in `internal/service`, persistence in `internal/repository`, and middleware in `internal/middleware`. Database migrations belong in `internal/http/database/migrations`. Generated Swagger files are in `docs`; regenerate them rather than editing by hand. Order-service protobuf/client types are under `third_party/booknest-order-service`.

## Build, Test, and Development Commands

- `go mod download`: install Go module dependencies.
- `make run`: run the API locally with `go run .`; expects a root `.env`.
- `go test ./...`: run the full test suite.
- `go test ./internal/middleware/...`: run one package.
- `go test -run TestName -count=1 ./internal/path/...`: run a specific test without cache.
- `make migrate-up`: apply SQL migrations using `DB_URL`.
- `make migrate-new name=<migration_name>`: create paired migration files.
- `make doc`: regenerate Swagger with `swag init --parseDependency --parseInternal`.
- `docker compose up --build`: run API, PostgreSQL/pgvector, and Redis locally.

## Coding Style & Naming Conventions

Use standard Go formatting: run `gofmt` on changed `.go` files and keep imports organized by `goimports` or the editor. Package names are lowercase and concise. Test files use `*_test.go`; tests should read as behavior-focused `Test...` functions. Keep layered boundaries clear: controllers validate/translate HTTP, services hold business rules, repositories perform database access, and domain types stay framework-light.

## Testing Guidelines

The project uses Go’s `testing` package with `testify`, plus `pgxmock` and SQLite/GORM helpers in repository tests. Add or update tests beside the changed package. Prefer focused unit tests for services and middleware, and use controller/API tests for routing, auth, and response behavior. Run `go test ./...` before opening a PR.

## Commit & Pull Request Guidelines

Recent history mostly uses Conventional Commits, especially `feat: ...`; follow that style, for example `fix: handle missing AI provider config`. Keep commits scoped and descriptive. PRs should include a concise summary, test results, migration notes, and linked issues when applicable. Include screenshots only for visible API documentation or client-facing behavior changes.

## Security & Configuration Tips

Do not commit real `.env` files, credentials, tokens, or AWS/OpenAI secrets. Local configuration is documented in `README.md`; production-sensitive behavior includes Redis TLS when `ENV=production`, S3 uploads, SES/SNS notifications, JWT secrets, and AI provider credentials. Treat generated files and migrations carefully: update Swagger after API annotation changes, and never rewrite applied migrations without coordination.
