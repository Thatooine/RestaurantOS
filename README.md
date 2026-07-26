# The Shire Shack

## Introduction

A restaurant-management REST API built with Go and Gin. Restaurant owners can register an establishment and manage dishes; authenticated customers can browse, search, and rate dishes.

The backend uses Gin `v1.12.0`. The completed [Gorilla Mux to Gin migration plan](docs/gin-migration-plan.md) records the compatibility decisions and verification criteria.

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.25.7+
- [Docker](https://www.docker.com/) and Docker Compose
- `curl` for the startup health check

### One-Command Start

To start the infrastructure, bootstrap the database, and run the API server:

```shell
./scripts/start.sh
```

This starts the Docker Compose services, replaces the local seed data, creates indexes, and launches the Go API on `:8080`. Press `Ctrl+C` to stop the API and the Compose services.

---

If you prefer to start each component manually, follow the steps below.

### 1. Start Infrastructure

From the project root, bring up the infrastructure:

```shell
docker compose up -d
```

This starts:

- MongoDB on `localhost:27017`, running as a single-node replica set (`rs0`) so multi-document transactions work. The replica set is initiated automatically by the container's healthcheck a few seconds after startup; clients connect with `?directConnection=true`.
- Redis on `localhost:6379`
- RabbitMQ on `localhost:5672` (management UI at `localhost:15672`, guest/guest)
- Prometheus on `localhost:9090`
- Loki on `localhost:3100`
- Grafana on `localhost:3001` (admin/admin; anonymous viewer access and Prometheus/Loki datasources are provisioned)
- Grafana Alloy on `localhost:12345`, with OTLP gRPC on `localhost:4317` and OTLP HTTP on `localhost:4318`

### 2. Bootstrap the Database

Seed the database with a root user, a sample restaurant, and sample dishes, and create all indexes:

```shell
go run ./cmd/bootstrap
```

The bootstrap command clears the application collections before seeding them. Do not point the local configuration at data you need to retain.

It creates this **root user** for testing:

| Field | Value |
|-------|-------|
| Email | `root+user@gmail.com` |
| Password | `abc123` |
| ID | `00000000-0000-0000-0000-000000000000` |
| Roles | Admin, RestaurantOwner |

Set `MONGO_URI` to bootstrap a non-default MongoDB instance. Set `BOOTSTRAP_ROOT_PASSWORD` to override the local default password shown above.

Expected output includes:

```
cleared users collection
root user created: root+user@gmail.com (00000000-0000-0000-0000-000000000000)
cleared restaurants collection
seed restaurant created: The Dancing Pony (00000000-0000-0000-0000-000000000001)
cleared dishes collection
dish created: Lembas Bread (00000000-0000-0000-0000-000000000101)
dish created: Shire Mushroom Stew (00000000-0000-0000-0000-000000000102)
dish created: Second Breakfast Platter (00000000-0000-0000-0000-000000000103)
index ensured: users.id (unique=true)
index ensured: users.email (unique=true)
...
```

### 3. Start the API Server

```shell
go run ./cmd/app
```

The API starts on `http://localhost:8080`. Verify with:

```shell
curl http://localhost:8080/health
```

## How It Works

### Architecture

```
HTTP Client
    |
    v
Go HTTP Server (Gin)
    |
    |-- Logger Middleware
    |-- Prometheus Metrics Middleware
    |-- Auth Middleware (app-issued JWT)
    |-- User Rate Limiter Middleware (Redis token bucket, per-user)
    |
    |-- IP Rate Limiter Middleware (Redis token bucket, per-IP, login only)
    |
    |-- REST Adaptors (HTTP <-> Service translation)
    |       |
    |       v
    |-- Service Layer (business logic, interfaces in pkg/, impl in internal/)
    |       |
    |       v
    |-- MongoDB Store (data access)
    |
    v
MongoDB          Redis
(data)           (rate limiting state)
```

### API Endpoints

**Unauthenticated:**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics scrape endpoint |
| POST | `/api/v1/auth/login` | Login with email/password (IP rate-limited) |
| POST | `/api/v1/auth/register` | Register with email/password |

**Authenticated** (require JWT via `Authorization: Bearer` header or `access_token` cookie):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/restaurants` | List all restaurants |
| GET | `/api/v1/restaurants/mine` | Get my restaurant |
| GET | `/api/v1/restaurants/search?q=` | Search restaurants |
| GET | `/api/v1/restaurants/{id}` | Get restaurant by ID |
| POST | `/api/v1/restaurants/register` | Register a new restaurant |
| GET | `/api/v1/dishes` | List dishes (optionally `?restaurant_id=`) |
| GET | `/api/v1/dishes/search?q=` | Search dishes |
| GET | `/api/v1/dishes/{id}` | Get dish by ID |
| POST | `/api/v1/dishes` | Create dish (owner only) |
| PUT | `/api/v1/dishes/{id}` | Update dish (owner only) |
| DELETE | `/api/v1/dishes/{id}` | Delete dish (owner only) |
| GET | `/api/v1/dishes/{id}/ratings` | List ratings for a dish |
| POST | `/api/v1/dishes/{id}/ratings` | Submit a rating |
| GET | `/api/v1/users` | List users |
| GET | `/api/v1/users/search?q=` | Search users |
| GET | `/api/v1/users/{email}` | Get user by email |

### Authentication

- The API authenticates users by email/password and issues its own JWT
- The token can be provided in two ways:
  - **Header:** `Authorization: Bearer <token>`; the login endpoint returns the token in its JSON body
  - **Cookie:** `access_token`; the registration endpoint sets it as a `Secure`, `HttpOnly`, `SameSite=Strict` cookie
- All authenticated routes validate the JWT via middleware
- The middleware checks the `Authorization` header first, then falls back to the cookie

Because the registration cookie is `Secure`, most clients will not send it over local plain HTTP. For local development, log in and use the returned bearer token.

### Authorization

- Users start with the `Customer` role
- Registering a restaurant promotes the user to `RestaurantOwner`
- Dish create/update/delete verifies the caller is the owner of the target restaurant

### Rate Limiting

Two layers of rate limiting, both Redis-backed token buckets (via `mennanov/limiters`):

**User Rate Limiter** (`UserRateLimiterMiddleware`)

- Per-user, keyed by UserID from the JWT
- Applied to all authenticated routes
- Default: 20 requests burst, 1 token/second refill

**IP Rate Limiter** (`IpRateLimiterMiddleware`)

- Per-IP, keyed by client IP address
- Applied to `/api/v1/auth/login` only
- Default: 5 requests burst, 1 token/minute refill
- Protects against brute-force login attempts

### Observability

The API exposes Prometheus metrics at `/metrics`, recorded by middleware in `pkg/metrics`:

| Metric | Type | Labels |
|--------|------|--------|
| `http_requests_total` | Counter | `method`, `route`, `status` |
| `http_request_duration_seconds` | Histogram | `method`, `route` |
| `http_requests_in_flight` | Gauge | — |

The `route` label uses a canonical path template (for example, `/api/v1/dishes/{id}`) to keep cardinality bounded. Gin's native `:id` and `:email` parameters are normalized to the canonical brace format so existing Prometheus series remain backward compatible.

Prometheus scrapes the API every 15 seconds using `observability/prometheus/prometheus.yml`; targets are visible at <http://localhost:9090/targets>. Grafana is available at <http://localhost:3001>.

## Running Tests

Run the unit tests and compile every package without starting external services:

```shell
go test -short ./...
```

For integration tests, start and bootstrap the local stack in one terminal:

```shell
./scripts/start.sh
```

Then run the API integration suite in another terminal:

```shell
go test ./tests/integration -v
```

## Project Structure

```
cmd/
  app/              # API server entry point and wiring
  bootstrap/        # Database seed and index creation
docs/
  gin-migration-plan.md
internal/pkg/       # Private service implementations
  authentication/   # JWT issuance & validation, email/password auth
  rateLimiting/     # Redis rate limiter implementation
  restaurants/      # Restaurant, dish, rating services
  users/            # User services
pkg/                # Public interfaces and adaptors
  authentication/   # Auth interfaces, middleware, REST adaptors
  errs/             # Sentinel errors (ErrNotFound, ErrConflict, ErrForbidden)
  logger/           # Logging middleware
  metrics/          # Prometheus middleware and /metrics handler
  mongo/            # MongoDB client, store abstraction, Storer interface
  rateLimiting/     # Rate limiter interfaces, middleware
  restaurants/      # Restaurant/dish/rating interfaces, REST adaptors
  users/            # User interfaces, REST adaptors
api/                # Postman collection
observability/      # Prometheus, Grafana, Loki, and Alloy configuration
scripts/            # Local stack orchestration
tests/integration/  # Live API and Redis-backed integration tests
```
