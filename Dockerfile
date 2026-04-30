# ---------- BUILD STAGE ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
# The main module uses a local replace, so the replacement module must exist
# before `go mod download` runs inside the image.
COPY third_party/booknest-order-service ./third_party/booknest-order-service
RUN go mod download

COPY . .
RUN go build -o app

# ---------- RUN STAGE ----------
FROM alpine:3.20

WORKDIR /app

# Copy Go binary
COPY --from=builder /app/app .

# Copy migrations to a SIMPLE runtime path
COPY --from=builder /app/internal/http/database/migrations ./migrations

# Copy entrypoint script
COPY entrypoint.sh .

# Install migrate
RUN apk add --no-cache curl ca-certificates tar file \
 && curl -fL https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-ard64.tar.gz \
    -o /tmp/migrate.tar.gz \
 && tar -xzf /tmp/migrate.tar.gz -C /tmp \
 && mv /tmp/migrate /usr/local/bin/migrate \
 && chmod +x /usr/local/bin/migrate \
 && file /usr/local/bin/migrate && chmod +x entrypoint.sh


EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
