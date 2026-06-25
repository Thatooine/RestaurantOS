---
name: scaffold-service
description: Given a domain SERVICE or CAPABILITY interface (anything whose name ends in "Service", reads as a capability like UserRegistration / RestaurantRegistration / RatingSubmitter, or that the user explicitly calls a "service" or "capability"), scaffold its implementation, its REST adaptor, its hand-written mock, its Validate() methods, and the pure-Go unit tests. Use when adding a domain service that orchestrates one or more repositories. NOT for plain entity persistence — that is scaffold-crud-repository. OUT OF SCOPE — does NOT create the underlying repository, and does NOT wire the adaptor into cmd/app.
---

# Scaffold a domain service / capability

A **service** (a.k.a. capability) is not a repository. A repository persists one entity (MongoDB, `mongo.Store`, `bson` filters). A service orchestrates the **domain policy the repository deliberately omits** — defaulting fields, pinning the owner to the caller, composing several repository / external-API calls into one operation — over the `pkg` repository *interfaces*. It touches **no `mongo.Store` directly** and contains **no `bson`**.

This project is **MongoDB-backed** and exposes everything over **REST** (gorilla/mux), not JSON-RPC. Module is `github.com/bash/the-dancing-pony-v2-rnyfbr`.

Trigger this skill when the target interface's name ends in `Service`, reads as a capability (`UserRegistration`, `RestaurantRegistration`, `RatingSubmitter`…), or the user explicitly says "service" / "capability". For plain entity CRUD persistence use `scaffold-crud-repository` instead.

The canonical references to mirror in every detail:
- **`UserService`** — the minimal read service that just delegates to one repository. Files: `pkg/users/userService.go`, `pkg/users/userServiceRESTAdaptor.go`, `internal/pkg/users/userServiceImpl.go`. (Its DTOs are shared with `UserRepository`, declared in `userRepository.go`.)
- **`UserRegistration`** — the multi-collaborator service that composes an external API (Firebase), the access-token creator, and the user repository. Files: `pkg/users/userRegistrationService.go`, `pkg/users/userRegistrationRESTAdaptor.go`, `internal/pkg/users/userRegistrationServiceImpl.go` + `..._test.go`.
- **`RestaurantRegistration`** — the owner-pinning write adaptor (pulls the login claim, pins `UserID` to it). Files: `pkg/restaurants/restaurantRegistrationService.go`, `restaurantRegistrationRESTAdaptor.go`, `internal/pkg/restaurants/restaurantRegistrationServiceImpl.go`.

## What this skill produces (and what it does NOT)

Produces (for service `<Svc>` in domain `<domain>`):

| File | Layer |
|------|-------|
| `pkg/<domain>/<svc>Service.go` | interface + request/response DTOs (only if the interface doesn't already exist) |
| `pkg/<domain>/<svc>Validations.go` | `Validate()` per request |
| `pkg/<domain>/<svc>Validations_test.go` | pure-Go table-driven validation tests |
| `pkg/<domain>/<svc>RESTAdaptor.go` | the wire layer (gorilla/mux HTTP handlers) |
| `internal/pkg/<domain>/<svc>ServiceImpl.go` | the orchestration impl |
| `internal/pkg/<domain>/<svc>ServiceImpl_test.go` | mock-first impl tests (in-package) |
| `internal/pkg/<domain>/<svc>Mock.go` | hand-written mock of the service interface (if a caller needs it) |

**Out of scope — do NOT create these:**
- the underlying repository (use `scaffold-crud-repository`); this skill **consumes** repository interfaces, it does not create them.
- the wiring in `cmd/app/serviceProvider.go` (the `ServiceProviders` struct field + the `New…` call in `NewServiceProviders`) and the route registration in `cmd/app/setupAPIServer.go` — a brand-new adaptor MUST be registered there or it is unreachable.

State this boundary back to the user when you finish.

## Step 0 — read the interface and decide three things

If the service interface already exists, read it. Otherwise define it (Step 1). Either way determine:

1. **Names.** Domain package `<domain>`; service type `<Svc>` (e.g. `RestaurantRegistration`); the method(s) it exposes. The impl is `<Svc>ServiceImpl` with `New<Svc>ServiceImpl(...)`; the adaptor is `<Svc>RESTAdaptor` with `New<Svc>RESTAdaptor(svc <Svc>)`.
2. **Collaborators.** Which `pkg/<domain>.XxxRepository` interfaces (and which methods) does the policy need? Which external collaborators (a `*firebase.App`, `authentication.AccessTokenCreatorService`)? The impl takes these as constructor params — always the `pkg` *interfaces* for repositories (so it is mock-testable), never the `internal` impls. Repositories must already exist; if one doesn't, stop and point at `scaffold-crud-repository`.
3. **HTTP shape.** A read (GET, path/query params → response) like `UserService`, or a write (POST/PUT, JSON body, owner pinned to the claim) like `RestaurantRegistration`? This decides the adaptor in Step 4.

There is **no SQL transaction manager** in this project. Mongo writes are not wrapped in a unit of work. When a service makes several writes that could partially fail, the pattern is **idempotent recovery**, not rollback: `UserRegistration.getOrCreateUser` re-reads by key first and only creates if absent, so a retry after a partial failure converges. Copy that shape — do not invent a `TxManager`.

## Step 1 — interface + DTOs (`pkg/<domain>/<svc>Service.go`, only if absent)

Mirror `userService.go` / `userRegistrationService.go`: a doc comment that names the policy the service owns (and, for a read service, that it is the REST-facing port delegating to the repository); the interface; and `<Method>Request` / `<Method>Response` structs. A thin read service can **reuse the repository DTOs** rather than declaring new ones (as `UserService` does).

**Ownership is the load-bearing rule:** if the operation acts on user-owned data, the *service* request DTO carries `UserID string`, and it is **always the calling principal, filled by the adaptor from the verified login claim, never from the wire** (see `RegisterRestaurantRequest.UserID`). Optional inputs the service defaults are documented as optional.

## Step 2 — validations + test (`pkg/<domain>/<svc>Validations.go` + `_test.go`)

`func (r *<Method>Request) Validate() error` in the `var reasons []string` / `strings.Join` style (`fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))`). Require only what the service genuinely needs. Add the table-driven, in-package, no-DB test per the `go-unit-tests` skill: first row valid, one failing row per required field. (A pure pass-through service like `UserService`, whose DTOs and validation already live with the repository, needs no new validation file.)

## Step 3 — impl (`internal/pkg/<domain>/<svc>ServiceImpl.go`)

Mirror `userServiceImpl.go` (thin delegate) or `userRegistrationServiceImpl.go` (multi-collaborator orchestration).

- `package <domain>` under `internal/pkg`; import the port as `"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/<domain>"`.
- Struct holds the collaborator interfaces / external clients; `New<Svc>ServiceImpl(...)` injects them. Add `var _ <domain>.<Svc> = &<Svc>ServiceImpl{}`.
- If the service has its own `Validate()`-bearing requests, **first line of every method:** `request.Validate()` → log + wrap `fmt.Errorf("invalid request for <Method>: %w", err)`. A pure delegate that forwards repository DTOs lets the repository validate.
- Apply the domain policy here: substitute defaults for blank optional fields; **pin the owner to `request.UserID`** (already the caller) when constructing the repository request — never trust a wire-supplied owner; stamp `time.Now()...` for time fields.
- For multi-step writes use the **get-or-create / re-read-then-write** idempotency pattern (see `getOrCreateUser`); there is no transaction.
- **Preserve sentinels:** wrap collaborator errors with `%w` (`fmt.Errorf("could not <do>: %w", err)`) so `errs.ErrNotFound` / `errs.ErrAlreadyExists` survive for callers/tests.

## Step 4 — REST adaptor (`pkg/<domain>/<svc>RESTAdaptor.go`)

Mirror `userServiceRESTAdaptor.go` (read) or `restaurantRegistrationRESTAdaptor.go` (write).

- Struct wraps the service **interface**; `New<Svc>RESTAdaptor(svc <Svc>)`.
- One method per operation with the **net/http handler signature**: `func (a *<Svc>RESTAdaptor) <Method>(w http.ResponseWriter, r *http.Request)`. There is no `Name()` and no gorilla/rpc — routes are registered by hand in `setupAPIServer.go`.
- Declare wire DTOs (`<Method>RESTRequest` / `<Method>RESTResponse`) with `json:"snake_or_camelCase"` tags as needed. **The owner is NOT a wire field.**
- **Reads:** pull path vars with `mux.Vars(r)["..."]` and query params with `r.URL.Query().Get(...)` + `strconv.Atoi`; default `limit` to 20 when 0 (see `ListUsers`).
- **Writes / owner-scoped:** pull the claim and **fail closed** before doing anything:
  ```go
  claim, ok := authentication.LoginClaimFromContext(r.Context())
  if !ok {
      log.Ctx(r.Context()).Warn().Msg("no login claim in context")
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(http.StatusUnauthorized)
      json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
      return
  }
  ```
  Decode the JSON body, then build the service request with `UserID: claim.UserID` (pin to caller, never the body).
- Map service errors with `errs.WriteHTTPError(w, err)` (it translates `errs.ErrNotFound` etc. to opaque HTTP statuses — never leak existence). On success set `Content-Type: application/json`, `w.WriteHeader(http.StatusOK/Created)`, and encode the response DTO.

## Step 5 — mocks

1. **The service's own mock** — `internal/pkg/<domain>/<svc>Mock.go`, only if a *caller* (another service, an adaptor test) needs to stand it in. Hand-written function-field mock, identical shape to `accessTokenCreatorServiceMock.go` / `userRepositoryMock.go`: `var _ pkg<Domain>.<Svc> = &<Svc>Mock{}`, one `<Method>Fn func(ctx context.Context, request ...) (..., error)` field per method, each method delegating straight to its `Fn` field. **No `t *testing.T` parameter, no mutex** unless a test needs call-count assertions.
2. **Collaborators are already mocked.** The impl test reuses the existing repository mocks (`<Entity>RepositoryMock` with `<Method>Fn` fields) and external-service mocks (e.g. `AccessTokenCreatorServiceMock`).

## Step 6 — impl test (`internal/pkg/<domain>/<svc>ServiceImpl_test.go`)

Mock-first, **`package <domain>` in-package** under `internal/pkg` (the mocks live in the same internal package, so no import cycle — there is no need for an external `_test` package). Mirror `userRegistrationServiceImpl_test.go`. Mandatory cases (happy-path-AND-error-case rule):
- **Happy path** — returns the right value AND, as a **spy** on a collaborator mock, asserts what the service passed it: owner pinned to the request `UserID`, defaults applied.
- **Default/derived field** — a blank optional input is replaced before the collaborator is called.
- **Collaborator error** — a repo / external mock returning `errs.ErrNotFound` (etc.) surfaces from the service and `errors.Is` still matches the sentinel through the wrap; the response is `nil`.
- **Fail-closed / not-called path** — where the policy must short-circuit, put `t.Fatal(...)` *inside* the collaborator's `Fn` to prove it is never reached (see `CreateUserFn` asserting it is not called when the user already exists).

## Finish

1. `go build ./... && go vet ./... && go test ./pkg/<domain>/... ./internal/pkg/<domain>/...` — all must pass.
2. Restate the boundary: **the service is implemented, wired to its repositories, and exposed via a REST adaptor type — but it is NOT yet reachable over the wire.** Still required: add a field to `ServiceProviders` and a `New<Svc>ServiceImpl(...)` call in `cmd/app/serviceProvider.go`, and register the adaptor's handlers with the router in `cmd/app/setupAPIServer.go` (`api.HandleFunc("/...", adaptor.<Method>).Methods(...)`). Without both, the endpoint does not exist.