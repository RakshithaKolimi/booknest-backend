# BookNest Testing Guide

This guide records the testing conventions already used in `booknest-backend` and sets the expected standard for future work. It is based on the current repository structure, README, Makefile, Docker setup, CI workflow, production code, and existing tests.

## Existing Repository Standards

BookNest is a Go REST API with a layered architecture:

- `internal/domain`: domain models and service/repository/controller interfaces.
- `internal/http/controller`: Gin HTTP handlers and route registration.
- `internal/http/api`: versioned API routers.
- `internal/service`: business logic and cross-repository orchestration.
- `internal/repository`: PostgreSQL, GORM, pgx, Squirrel, and pgvector access.
- `internal/middleware`: JWT auth, roles, CORS, rate limiting, logging, recovery, and Swagger auth.
- `internal/ai`, `internal/pkg/ai`, `internal/service/ai_service`, `internal/service/book_embedding_service`: OpenAI-compatible chat, embeddings, semantic search, and recommendations.
- `docs`: generated Swagger files.
- `third_party/booknest-order-service`: local gRPC/protobuf replacement module.

Tests are package-local and use Go’s standard `testing` package. `testify/require` is used mostly for setup and repository assertions. Test files use the `*_test.go` suffix and sit beside the package they validate.

Examples:

```text
internal/http/controller/user_controller_test.go
internal/service/ai_service/ai_service_test.go
internal/repository/cart_repository_test.go
internal/middleware/jwt_auth_test.go
main_test.go
docs/docs_test.go
```

## Running Tests

Run the full suite:

```bash
go test ./...
```

Run one package:

```bash
go test ./internal/http/controller
```

Run one test without cache:

```bash
go test -run TestLogin_Success -count=1 ./internal/http/controller
```

Generate coverage, matching CI:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Local development commands also include:

```bash
make run
make migrate-up
make migrate-new name=<migration_name>
make doc
docker compose up --build
```

## CI/CD Validation

GitHub Actions runs on pushes to `main`. The `test` job checks out the repo, installs Go, downloads modules, runs:

```bash
go test ./... -coverprofile=coverage.out
```

It then prints coverage, uploads `coverage.out` as an artifact, and sends coverage to Codecov. The Docker job runs only after tests pass, builds the API image, and pushes `ghcr.io/<owner>/booknest-platform:latest`.

Current deviation: `go.mod` and README specify Go `1.24`, while `.github/workflows/ci.yml` uses Go `1.22`. Align CI to Go `1.24` to match the project contract.

## Test Naming Conventions

Use behavior-focused names:

```go
func TestLogin_Success(t *testing.T) {}
func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {}
func TestBookRepo_CreateAndUpdateWithRelations(t *testing.T) {}
```

Current patterns include:

- `TestFunction_Condition`
- `TestServiceBehavior`
- `TestRepo_ActionAndOutcome`
- Table-driven subtests when multiple cases share setup.

Keep names explicit enough that a failing test explains the broken behavior.

## Dependency Injection Patterns

Production code uses constructor injection through domain interfaces:

```go
bookRepo := repository.NewBookRepository(gormdb, sqlDB)
aiService := ai_service.NewAIService(aiProvider, bookEmbeddingRepo, orderRepo)
bookService := book_service.NewBookService(bookRepo, categoryRepo, embeddingService, bookEmbeddingRepo, orderRepo, aiService)
bookController := controller.NewBookController(bookService)
```

Tests should preserve this style. Inject fake repositories, providers, notification senders, or services instead of reaching into global state. When global state is unavoidable, restore it with `t.Cleanup`.

Examples already in use:

- `main_test.go` overrides `connectGORM`.
- `internal/http/database/database_test.go` overrides pool creation and ping functions.
- Controller tests inject mock services into `New...Controller`.
- Service tests inject mock repositories and providers.

## Mocking Strategies

Prefer small package-local mocks with function fields:

```go
type mockProvider struct {
	replies []string
	err     error
}

func (m *mockProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	reply := m.replies[0]
	m.replies = m.replies[1:]
	return reply, nil
}
```

Use existing tools where the repository already does:

- `httptest.NewRequest` and `httptest.NewRecorder` for controllers and middleware.
- `gin.SetMode(gin.TestMode)` for Gin tests.
- `pgxmock` for pgx/Squirrel SQL expectations.
- In-memory SQLite/GORM via `setupTestDB(t, models...)` for model persistence tests.
- `t.Setenv` for environment-dependent behavior.
- `testify/require` when setup failure should stop the test immediately.

Do not use real AWS, OpenAI, Redis, gRPC, or persistent PostgreSQL dependencies in unit tests.

## Layer-Specific Testing Requirements

### Controllers

Controller tests must validate HTTP status codes, JSON response shapes, route registration, auth behavior, and input sanitization. Use Gin test mode and package-local mock services. For authenticated routes, configure JWT test keys and include a token or set context values through helper routes.

Required coverage for changed controller behavior:

- Success response.
- Invalid JSON or invalid input.
- Missing/invalid authentication when route requires JWT.
- Service-layer error mapping to the expected HTTP status.

### Services

Service tests must focus on business rules, dependency interactions, and error propagation. Mock repositories and external providers. For BookNest-specific flows, cover stock validation, order authorization, notification preferences, purchase-gated reviews, AI fallback behavior, semantic search, recommendations, and embedding refresh behavior.

Required coverage for changed service behavior:

- Happy path.
- Validation error.
- Repository/provider error.
- Important side effect, such as stock decrement, notification send, token write, or embedding upsert.

### Repositories

Repository tests should use `pgxmock` for SQL shape and arguments, or in-memory SQLite/GORM when validating GORM model behavior. Always assert `mock.ExpectationsWereMet()` for pgxmock tests. For Squirrel-built SQL, assert durable query fragments rather than brittle full strings unless the exact SQL is the behavior under test.

BookNest repository tests should cover:

- UUID handling.
- Soft-delete filters.
- Pagination and cursor behavior.
- Transactions and rollback paths.
- pgvector search query behavior.
- Relation creation and replacement for books, authors, publishers, and categories.

### Middleware

Middleware tests should use `httptest` and small Gin routes. Cover success and failure paths for JWT auth, role checks, rate limiting, Swagger basic auth, security headers, and error handling.

### AI and Embeddings

AI tests must avoid network calls. Use fake providers for `Generate` and `Embed`. Cover:

- Empty prompt/input validation.
- Provider unavailable errors.
- Intent detection response parsing.
- Chat fallback from `message` to legacy `prompt`.
- Semantic search and recommendation reference behavior.
- Embedding vector length and repository upsert behavior.
- Best-effort behavior in book creation/update when AI is unavailable.

### Configuration and Startup

Use `t.Setenv` for environment variables. Startup tests should keep database, Redis, notification, and AI dependencies fake or optional. Cover env parsing errors, defaults, and compatibility variables such as `ORDER_SERVICE_MODE`, `USE_ORDER_MICROSERVICE`, `BOOKNEST_WEB_URL`, and `FRONTEND_URL`.

## Error Handling Standards

The codebase uses plain `errors.New`, wrapped `fmt.Errorf("context: %w", err)`, sentinel errors, and `errors.Is`. Continue those patterns.

Examples:

- `ai_service.ErrProviderUnavailable`
- `domain.ErrReviewRequiresPurchase`
- `fmt.Errorf("parse USE_ORDER_MICROSERVICE: %w", err)`

Controller tests should assert the HTTP status and error message when behavior depends on error mapping. Service tests should use `errors.Is` for sentinel errors.

## Logging and Concurrency Standards

Logging currently uses both `log` and `log/slog`. Middleware logs method, path, status, and latency with `log.Printf`; startup and background AI embedding work often use `slog`.

Concurrency exists in:

- HTTP server startup and graceful shutdown.
- Startup embedding sync.
- Background embedding refresh with `sync.Map` deduplication.
- User notification goroutines.
- In-memory rate limiter mutex usage.

When changing concurrent code, tests should avoid sleeps where possible. Prefer controllable fakes, channels, short contexts, and assertions that background work is scheduled or deduplicated.

## Required Practices for Future Development

- Add or update tests with every production behavior change.
- Keep tests package-local unless black-box behavior specifically requires an external test package.
- Use existing constructors and interfaces for dependency injection.
- Use fakes/mocks instead of real external services.
- Use `t.Setenv` for environment changes.
- Run `go test ./...` before opening a PR when feasible.
- Regenerate Swagger with `make doc` after API annotation changes, then keep `docs/*` tests passing.
- Do not weaken assertions, delete tests, or bypass errors simply to make a build pass.

## Recommended Improvements

These improvements extend current conventions without replacing them:

- Align CI Go version with `go.mod` (`1.24`) to prevent local/CI drift.
- Add a `make test` target that runs `go test ./... -coverprofile=coverage.out`, matching CI.
- Add focused tests for `internal/service/book_embedding_service`.
- Add tests for `internal/repository/book_embedding_repository`, especially pgvector query construction and validation.
- Add controller tests for role-only middleware paths where admin behavior changes.
- Add concurrency-focused tests for background embedding scheduling and deduplication.
- Add a coverage threshold in CI after baseline coverage is agreed.
- Consider consolidating duplicate AI provider packages (`internal/ai/provider` and `internal/pkg/ai`) or document why both exist.

# Repository Testing Audit

## Current Strengths

- Broad package coverage across controllers, services, repositories, middleware, startup, docs, and utilities.
- Tests are colocated with code and use standard Go tooling.
- Dependency injection through domain interfaces makes services and controllers easy to test.
- Existing mocks are small and readable.
- CI already runs the full Go suite with coverage before publishing Docker images.
- Environment-sensitive code commonly uses `t.Setenv`.
- SQL tests use `pgxmock`; GORM persistence tests use in-memory SQLite helpers.
- AI service tests avoid network calls and validate intent, embeddings, recommendations, and provider-unavailable paths.

## Current Weaknesses

- No `CONTRIBUTING.md` exists, so testing expectations were previously spread across README, Makefile, CI, and test files.
- CI uses Go `1.22` while the repository declares Go `1.24`.
- There is no `make test` command even though testing is central to CI.
- Coverage is collected but no minimum threshold is enforced.
- Some high-risk AI/embedding and background-concurrency paths have limited direct tests.
- Logging uses both `log` and `slog`; tests generally assert behavior, not log output.
- Some tests use large hand-written mocks, which can become noisy as interfaces grow.

## Missing Coverage Areas

- `internal/service/book_embedding_service/book_embedding_service.go`.
- `internal/repository/book_embedding_repository.go`.
- `internal/ai/embedding_text.go`.
- `internal/ai/provider/*` direct provider package behavior.
- Route-only files such as `*_routes.go` beyond registration smoke tests.
- Background embedding sync and scheduled refresh deduplication behavior.
- Docker entrypoint migration behavior.
- More direct tests for admin role and authorization edge cases.

## High-Risk Modules Lacking or Needing More Tests

- Book embeddings and pgvector semantic search: affects AI search and recommendations.
- Book service background jobs: concurrent, best-effort, and easy to regress silently.
- Remote order service: gRPC mapping and fallback behavior are business-critical.
- Notification setup and templates: external-provider configuration and customer communication.
- Auth, refresh token, verification, and password reset flows: security-sensitive.
- Docker/entrypoint migrations: deployment-sensitive and not validated by Go tests.

## Recommended Priorities

1. Align CI with Go `1.24` and add `make test`.
2. Add unit tests for `book_embedding_service`.
3. Add pgxmock or focused SQL tests for `book_embedding_repository.SearchNearestBooks`.
4. Add concurrency tests for embedding refresh scheduling and deduplication.
5. Expand auth/role tests around admin-only routes and token edge cases.
6. Add smoke validation for Docker entrypoint or document manual deployment checks.
7. Establish and enforce a coverage threshold once the above gaps are closed.

## Future AI Agent Instructions

Codex, Claude Code, Cline, Cursor, and other AI coding agents must follow these rules:

- Always read `TESTING.md` before generating or modifying code.
- Follow repository testing conventions instead of introducing unrelated frameworks.
- Generate tests alongside production code.
- Maintain consistency with package-local mocks, constructor injection, and existing helpers.
- Never bypass, delete, or weaken tests to make builds pass.
- Update tests when business logic, route behavior, env config, or error handling changes.
- Avoid real network calls and real secrets in generated tests.
- Prefer targeted tests first, then broader package or full-suite validation.
- If proposing a better practice, explain why it improves on the current repository pattern.
