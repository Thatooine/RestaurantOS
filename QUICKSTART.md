# The Orc Shack

## Introduction

A restaurant management platform built with Go (REST API) and React (SPA frontend). Restaurant owners can register their establishment, manage dishes, and collect customer ratings. Customers can browse, search, and rate dishes.

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+ and [Yarn](https://yarnpkg.com/) (`npm install -g yarn`)
- [Docker](https://www.docker.com/) and Docker Compose

### One-Command Start

To start the entire stack (infrastructure, bootstrap, API server, and UI) in one go:

```shell
./scripts/start.sh
```

This will start MongoDB, Redis, Prometheus, Grafana, seed the database, launch the Go API on `:8080`, and the React UI on `:5173`. Press `Ctrl+C` to stop everything.

---

If you prefer to start each component manually, follow the steps below.

### 1. Start Infrastructure

From the project root, bring up MongoDB and Redis:

```shell
docker-compose up -d
```

This starts:
- MongoDB on `localhost:27017`
- Redis on `localhost:6379`
- Prometheus on `localhost:9090`
- Grafana on `localhost:3000` (admin/admin; Prometheus datasource auto-provisioned)

### 2. Bootstrap the Database

Seed the database with a root user, a sample restaurant, and sample dishes, and create all indexes:

```shell
go run ./cmd/bootstrap
```

This creates a **root user** for testing purposes:

| Field | Value |
|-------|-------|
| Email | `root+user@gmail.com` |
| Password | `abc123` |
| ID | `00000000-0000-0000-0000-000000000000` |
| Roles | Admin, RestaurantOwner |

Expected output:
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

### 4. Start the Frontend

In a separate terminal:

```shell
cd web
yarn install
yarn dev
```

The frontend starts on `http://localhost:5173` and proxies API calls to the backend.

## How It Works

### Architecture

```
Browser (React SPA)
    |
    | /api/v1/* (proxied by Vite in dev)
    v
Go HTTP Server (gorilla/mux)
    |
    |-- Logger Middleware
    |-- Auth Middleware (JWT via Firebase)
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
| POST | `/api/v1/auth/firebase` | Login with Firebase token |
| POST | `/api/v1/auth/register` | Register with email/password |
| POST | `/api/v1/auth/register/firebase` | Register with Firebase token |

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

- Firebase handles identity (email/password and Google OAuth)
- The API issues its own JWT; the token can be provided in two ways:
  - **Cookie:** `access_token` (set automatically by the login/register endpoints as an HttpOnly cookie)
  - **Header:** `Authorization: Bearer <token>`
- All authenticated routes validate the JWT via middleware
- The middleware checks the `Authorization` header first, then falls back to the cookie

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

The `route` label uses the gorilla/mux path template (e.g. `/api/v1/dishes/{id}`) to keep cardinality bounded.

Prometheus scrapes the API every 15s (config in `observability/prometheus/prometheus.yml`); targets are visible at http://localhost:9090/targets. Grafana (http://localhost:3000) has the Prometheus datasource auto-provisioned via `observability/grafana/provisioning/`.

## Running Tests

```shell
go test ./... -v
```

## Project Structure

```
cmd/
  app/              # API server entry point and wiring
  bootstrap/        # Database seed and index creation
internal/pkg/       # Private service implementations
  authentication/   # Firebase + JWT auth
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
web/                # React frontend (Vite + TypeScript)
  src/api/          # API client layer
  src/components/   # UI components
  src/context/      # Auth context provider
  src/hooks/        # Custom React hooks
  src/pages/        # Page components
```
