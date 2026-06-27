#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

cleanup() {
    echo ""
    echo "Shutting down..."
    if [ -n "${UI_PID:-}" ]; then
        kill "$UI_PID" 2>/dev/null || true
        wait "$UI_PID" 2>/dev/null || true
    fi
    if [ -n "${APP_PID:-}" ]; then
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    docker compose down
    echo "Done."
}
trap cleanup EXIT INT TERM

echo "==> Starting MongoDB and Redis..."
docker compose up -d
echo "==> Waiting for services to be ready..."

# Wait for MongoDB replica set to elect a primary (transactions require a replica set).
# The compose healthcheck initiates the replica set; here we wait until it is writable.
until docker compose exec mongoDB mongosh --quiet --eval "db.hello().isWritablePrimary" 2>/dev/null | grep -q true; do
    sleep 1
done
echo "    MongoDB is ready (replica set primary elected)."

# Wait for Redis
until docker compose exec redis redis-cli ping 2>/dev/null | grep -q PONG; do
    sleep 1
done
echo "    Redis is ready."

echo "==> Running bootstrap (seed data + indexes)..."
go run ./cmd/bootstrap

echo "==> Starting application server..."
go run ./cmd/app &
APP_PID=$!

# Wait for the app to start accepting connections
echo "==> Waiting for app to be ready on :8080..."
until curl -sf http://localhost:8080/health >/dev/null 2>&1; do
    sleep 1
done
echo "==> App is ready at http://localhost:8080"
cd "$ROOT_DIR"

echo ""
echo "==> Stack is running:"
echo "    API: http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop."
wait "$APP_PID"
