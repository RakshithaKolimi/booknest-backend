#!/bin/sh
set -e

echo "Running database migrations..."

migrate \
  -path /app/migrations \
  -database "${DB_URL}" \
  up

echo "Starting application..."
exec ./app
