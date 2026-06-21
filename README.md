# BookNest Backend

Go backend for the BookNest bookstore platform. It exposes a versioned REST API with Gin, stores data in PostgreSQL with pgvector support, uses GORM and pgx for data access, supports Redis-backed rate limiting, and includes Swagger docs for local API exploration.

## Features

- Versioned API mounted at `/api/v1`
- JWT authentication with refresh tokens
- Email verification, password reset, and mobile OTP flows
- Book catalog, author, category, and publisher management
- Cart, checkout, payment confirmation, cancellation, and admin order views
- Purchase-gated book reviews
- Admin book-cover uploads to S3
- AI chat, summaries, categories, embeddings, semantic search, and recommendations
- Optional gRPC forwarding to `BookNest-OrderService`
- Swagger UI protected by basic auth

## Tech Stack

- Go `1.24`
- Gin
- PostgreSQL / pgvector
- GORM and pgx
- Redis
- AWS SES, SNS, and S3
- OpenAI-compatible AI provider
- golang-migrate
- swaggo

## Project Structure

- `main.go`: startup, environment loading, Swagger registration
- `main_setup.go`: dependency wiring, middleware, route mounting, HTTP server setup
- `internal/http/api/v1`: active API version registrar
- `internal/http/controller`: HTTP handlers and route registration
- `internal/service`: business logic
- `internal/repository`: PostgreSQL and GORM data access
- `internal/http/database/migrations`: SQL migrations
- `internal/middleware`: auth, CORS, rate limiting, logging, error handling
- `internal/pkg/storage`: S3 image upload helpers
- `internal/ai` and `internal/pkg/ai`: AI provider integration
- `docs`: generated Swagger assets
- `third_party/booknest-order-service`: local replacement for order-service protobuf/client types

## Prerequisites

- Go `1.24+`
- PostgreSQL with pgvector support
- Redis `7+` for local parity with the deployed rate limiter
- `migrate` CLI for SQL migrations
- `swag` CLI when regenerating Swagger docs

## Environment

Create a `.env` file in the repository root. Use local-only values and do not commit real credentials.

```env
ENV=development

DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=booknest
DB_NAME=booknest
DB_PORT=5432
DB_URL=postgres://postgres:booknest@localhost:5432/booknest?sslmode=disable

JWT_SECRET_V1=replace-me
JWT_REFRESH_V1=replace-me
JWT_SECRET_V0=
JWT_REFRESH_V0=

SWAGGER_USER=booknest
SWAGGER_PASSWORD=replace-me

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

API_HOST=localhost:8080
BOOKNEST_WEB_URL=http://localhost:3000
# FRONTEND_URL is also supported for backward compatibility.

ORDER_SERVICE_MODE=monolith
USE_ORDER_MICROSERVICE=false
ORDER_SERVICE_GRPC_ADDR=localhost:50051

SES_FROM_EMAIL=
EMAIL_FROM=
SES_REGION=ap-south-1
SES_ACCESS_KEY=
SES_SECRET_KEY=

AWS_REGION=ap-south-1
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_BUCKET_NAME=booknest-images-prod
AWS_S3_ACCESS_KEY_ID=
AWS_S3_SECRET_ACCESS_KEY=

AI_PROVIDER=openai
OPENAI_API_KEY=
OPENAI_CHAT_MODEL=gpt-5.4-nano
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_BASE_URL=
OPENAI_HTTP_TIMEOUT_SECONDS=30
```

Notes:

- Startup calls `godotenv.Load(".env")`, so a local `.env` file must exist.
- `DB_URL` is preferred; when it is empty, the app builds a DSN from the individual `DB_*` values.
- If `REDIS_ADDR` is empty, login rate limiting falls back to in-memory storage.
- In production, the Redis client enables TLS when `ENV=production`.
- CORS always allows `http://localhost:3000` and `http://localhost:5173`; `BOOKNEST_WEB_URL` and `FRONTEND_URL` add deployed frontend origins.
- Notification setup requires SES and/or SNS credentials. If neither provider is configured, server setup fails.
- S3 book-cover uploads use `AWS_REGION`, `AWS_BUCKET_NAME`, `AWS_S3_ACCESS_KEY_ID`, and `AWS_S3_SECRET_ACCESS_KEY`.
- SMS notifications use `AWS_REGION`, `AWS_ACCESS_KEY_ID`, and `AWS_SECRET_ACCESS_KEY`.
- AI features require `OPENAI_API_KEY`; if the AI provider is not configured, AI-dependent features run best-effort where supported.

## Local Setup

1. Start PostgreSQL and create the `booknest` database.
2. Enable pgvector in the database if your local Postgres image does not include it by default.
3. Start Redis on `localhost:6379`.
4. Create the `.env` file.
5. Download dependencies and run migrations.

```bash
go mod download
make migrate-up
```

If you recently pulled new changes, rerun `make migrate-up` before starting the API.

## Run

```bash
make run
```

The API starts on `http://localhost:8080`.

Useful URLs:

- Health: `GET http://localhost:8080/health`
- API base: `http://localhost:8080/api/v1`
- AI health: `GET http://localhost:8080/api/v1/ai/health`
- Swagger UI: `http://localhost:8080/swagger/v1/index.html`

## Docker

The repository includes Docker files for running the API, PostgreSQL with pgvector, and Redis:

```bash
docker compose up --build
```

Before first use, make sure `.env` matches the container database settings. The compose stack exposes:

- API: `localhost:8080`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

## Common Commands

```bash
go test ./...
make run
make migrate-up
make migrate-down
make migrate-down-all
make migrate-version
make migrate-force VERSION=<n>
make migrate-new name=<migration_name>
make doc
```

## API Surface

Routes are mounted under `/api/v1` unless noted.

Auth:

- `POST /register`
- `POST /register-admin`
- `POST /login`
- `POST /refresh`
- `POST /forgot-password`
- `POST /reset-password`
- `POST /reset-password/confirm`
- `POST /verify-email`
- `POST /resend-email-verification`

Users:

- `GET /user/:id`
- `PUT /user/:id/preferences`
- `DELETE /user/:id`
- `POST /verify-mobile`
- `POST /resend-mobile-otp`

Books:

- `GET /books`
- `GET /books/search`
- `GET /books/semantic-search`
- `POST /books/filter`
- `GET /books/recommend`
- `GET /books/:id`
- `POST /books`
- `PUT /books/:id`
- `DELETE /books/:id`
- `POST /books/:id/summary`
- `POST /books/:id/categories`
- `POST /books/:id/embeddings`

Reviews:

- `GET /books/:id/reviews`
- `POST /books/:id/reviews`

Cart:

- `GET /cart`
- `POST /cart/items`
- `PUT /cart/items`
- `DELETE /cart/items/:book_id`
- `POST /cart/clear`

Orders:

- `GET /orders`
- `POST /orders/checkout`
- `POST /orders/confirm`
- `POST /orders/cancel`
- `GET /admin/orders`
- `PUT /admin/orders/status`

Catalog:

- `GET /authors`
- `POST /authors`
- `GET /authors/:id`
- `PUT /authors/:id`
- `DELETE /authors/:id`
- `GET /categories`
- `POST /categories`
- `GET /categories/:id`
- `PUT /categories/:id`
- `DELETE /categories/:id`
- `GET /publishers`
- `POST /publishers`
- `GET /publishers/:id`
- `PUT /publishers/:id`
- `PATCH /publishers/:id/status`
- `DELETE /publishers/:id`

Images and AI:

- `POST /images/upload`
- `GET /ai/health`
- `POST /ai/chat`

## Reviews

- `GET /books/:id/reviews` is public.
- `POST /books/:id/reviews` requires authentication.
- Users can create or update a review only after completing a purchase for that book.

## Image Uploads

- `POST /images/upload` accepts multipart form data with an `image` file field.
- The route requires an admin JWT.
- The response contains `{ "url": "..." }`; use that URL as `image_url` when creating or updating a book.

## Order Microservice Mode

Order handling defaults to the monolith implementation. To forward supported order operations to `BookNest-OrderService`:

1. Start the order service infrastructure.
2. Run the order service gRPC server on `localhost:50051`.
3. Set `ORDER_SERVICE_MODE=microservice` or `USE_ORDER_MICROSERVICE=true`.
4. Start this API.
5. Call `POST /api/v1/orders/checkout` and confirm the request is forwarded over gRPC.

When microservice mode is enabled, `ORDER_SERVICE_GRPC_ADDR` is required.

## Swagger

- Swagger is registered as the named instance `booknest-v1`.
- The documented base path is `/api/v1`.
- Access uses `SWAGGER_USER` and `SWAGGER_PASSWORD`.
- Regenerate docs after annotation changes:

```bash
make doc
```
