---
name: scaffold-crud-repository
description: Given a domain entity struct (e.g. pkg/users/user.go), scaffold its persistence layer — the UserRepository-style port interface + CRUD DTOs, the Validate() methods, the MongoDB-backed implementation, the hand-written mock, and the pure-Go validation unit tests. Use when adding persistence for a new entity. OUT OF SCOPE — does NOT create the exposed Service port, the REST adaptor, or the cmd/app wiring; the repository port is deliberately NOT reachable over REST.
---

# Scaffold a CRUD repository for an entity

Given one entity file `pkg/<domain>/<entity>.go` (a struct like `User` or `Dish`), generate the persistence layer following this repo's ports-and-adapters conventions. This project is **MongoDB-backed** (module `github.com/bash/the-dancing-pony-v2-rnyfbr`); `pkg/mongo` only exposes `NewClient`, so repositories talk to the driver (`go.mongodb.org/mongo-driver/v2/mongo`) **directly** — there is no `Store` wrapper. The canonical reference to mirror is the **users** domain (`pkg/users/userRepository.go` + `userRepositoryValidations.go` + `internal/pkg/users/userRepositoryMongoImpl.go` + `userRepositoryMock.go`); the **dishes** domain (`internal/pkg/restaurants/dishRepositoryMongoImpl.go`) is the second worked example.

## The layering this skill produces

```
REST  ──►  <Entity>Service (port)  ──►  <Entity>Repository (port)  ──►  *mongo.Client (driver)
          [exposed, see below]          [THIS SKILL — internal only]
```

The **repository** is the internal persistence port. It is never given a REST adaptor and is never put in a router. Anything that needs to reach an entity over the wire goes through a separate exposed `<Entity>Service` port + REST adaptor (see the out-of-scope note).

## What this skill produces (and what it does NOT)

Produces five files:

| File | Layer |
|------|-------|
| `pkg/<domain>/<entity>Repository.go` | interface + CRUD request/response DTOs |
| `pkg/<domain>/<entity>RepositoryValidations.go` | `Validate()` per request |
| `pkg/<domain>/<entity>RepositoryValidations_test.go` | pure-Go table-driven validation tests |
| `internal/pkg/<domain>/<entity>RepositoryMongoImpl.go` | MongoDB implementation |
| `internal/pkg/<domain>/<entity>RepositoryMock.go` | hand-written mock |

**Out of scope — do NOT create these:**
- the exposed `<Entity>Service` port (`pkg/<domain>/<entity>Service.go`)
- the REST adaptor (`pkg/<domain>/<entity>ServiceRESTAdaptor.go`)
- `cmd/app/serviceProvider.go` / `cmd/app/setupAPIServer.go` wiring

State this boundary back to the user when you finish, so they know the entity has persistence but is **not yet reachable over REST**.

## Step 0 — read the entity and decide names

Read `pkg/<domain>/<entity>.go`. From it derive:

1. **Names.** Package = `<domain>` (e.g. `users`). Entity = the struct name (e.g. `User`). Use the entity name to build method/DTO names; use a lowercase plural for the Mongo collection (`users`, `dishes`). The impl struct is `<Entity>RepositoryMongoImpl` holding a single `client *mongo.Client`, with constructor `New<Entity>RepositoryMongoImpl(client *mongo.Client)` — **the mongo client is the only dependency a repository may hold.**
2. **Methods.** Generate the operations the entity actually needs — don't force a rigid `Create/GetByID/List/Update/Delete` shape. The users repository, for example, is `CreateUser`, `GetUser` (by email), `ListUsers`, `SearchUsers`. Pick read keys that match how the entity is looked up (email, ID, owner), and a paginated `List` / `Search` returning `Total int64` alongside the slice.

There are **no SQL migrations** in this project. Collections are created lazily by Mongo; the only schema is the struct's `bson` tags. If a field must be uniquely indexed, that index is ensured from wiring, not the repository.

## Step 1 — interface + DTOs (`pkg/<domain>/<entity>Repository.go`)

Mirror `pkg/users/userRepository.go`. The interface is the **internal** persistence port — document that it is not exposed over REST. DTO rules:
- One `Xxx<Entity>Request` / `Xxx<Entity>Response` pair per method.
- Create carries the mutable input fields; its response carries the persisted `<Entity>`.
- Read-by-key carries the key (e.g. `Email string` / `<Entity>ID string`); its response carries the `<Entity>`.
- List carries `Offset int` / `Limit int`; its response carries `<Entity>s []<Entity>` and `Total int64`.
- Search carries `Query string` plus pagination; same response shape as List.

## Step 2 — validations (`pkg/<domain>/<entity>RepositoryValidations.go`)

One `func (r *Xxx<Entity>Request) Validate() error` per request, in the exact `var reasons []string` / `strings.Join` style of `userRepositoryValidations.go` (`fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))`):
- Non-empty check for every required ID/string field (e.g. `Name`, `Email`, `Query`).
- Numeric floors where they apply (`Limit < 0`, `Offset < 0`).

## Step 3 — validation tests (`pkg/<domain>/<entity>RepositoryValidations_test.go`)

Pure-Go, table-driven, **in-package** (`package <domain>`), no DB — follow the `go-unit-tests` skill. Per request: a `valid()` constructor returning a fully-populated request, then rows where the first is `{"valid", func(r *Req){}, false}` and every other row breaks exactly one field with `wantErr: true`. Cover the happy path AND one failing row per required field — the mandatory happy-path-AND-error-case rule.

## Step 4 — MongoDB impl (`internal/pkg/<domain>/<entity>RepositoryMongoImpl.go`)

Mirror `internal/pkg/users/userRepositoryMongoImpl.go` / `internal/pkg/restaurants/dishRepositoryMongoImpl.go` exactly. The repository drives the MongoDB driver directly — **its only field is `client *mongo.Client`.** Non-negotiable patterns:

- `package <domain>` under `internal/pkg`, importing the port as `"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/<domain>"` and the driver as `"go.mongodb.org/mongo-driver/v2/mongo"`, `".../mongo/options"`, `".../bson"`.
- Struct `<Entity>RepositoryMongoImpl{ client *mongo.Client }` + `New<Entity>RepositoryMongoImpl(client *mongo.Client)`.
- A `databaseName` const (`"restaurantos"`) is declared once per `internal/pkg/<domain>` package, plus a `collection() *mongo.Collection` helper returning `r.client.Database(databaseName).Collection("<plural>")`. Every method goes through `r.collection()`.
- Add a compile-time assertion: `var _ <domain>.<Entity>Repository = &<Entity>RepositoryMongoImpl{}` (drift breaks the build).
- **Every method starts with** `request.Validate()` — on failure `log.Ctx(ctx).Error().Err(err).Msg(...)` and return `fmt.Errorf("invalid request for <Method>: %w", err)`.
- **Create:** build the entity with a fresh ID — use the domain's helper if it exists (`users.NewID()`); otherwise `uuid.New().String()` (`github.com/google/uuid`). Then `r.collection().InsertOne(ctx, entity)`; wrap errors as `fmt.Errorf("<Method> failed: %w", err)`.
- **Get-by-key:** `r.collection().FindOne(ctx, bson.M{"<field>": key}).Decode(&entity)`. When not-found semantics matter, branch on `errors.Is(err, mongo.ErrNoDocuments)` → `fmt.Errorf("<entity> not found: %w", errs.ErrNotFound)` — never leak existence beyond that.
- **List / Search:** `CountDocuments` for the total, then `Find(ctx, filter, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))`, `defer cursor.Close(ctx)`, and `cursor.All(ctx, &slice)`. Factor the count+find+decode into a private `list(ctx, filter, offset, limit)` helper, as the users/dishes repos do. Search filters use a `$or` of `$regex` (`"$options": "i"`) clauses.
- **Update:** `r.collection().FindOneAndUpdate(ctx, bson.M{"id": id}, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&entity)`.
- **Delete:** `r.collection().DeleteOne(ctx, bson.M{"id": id})`; if `result.DeletedCount == 0`, return a not-found error.
- All driver errors are wrapped `fmt.Errorf("<Method> failed: %w", err)`.

Because a repository may depend on nothing but the mongo client, any cross-entity need (e.g. a service checking a user's role while writing a dish) is the **service's** job: the service composes the relevant repository ports. If a required lookup/mutation doesn't exist yet on the sibling repository, add it there (as `GetUserByID` / `UpdateUserRoles` were added to `UserRepository`) rather than reaching into another collection from the calling repository or service.

## Step 5 — mock (`internal/pkg/<domain>/<entity>RepositoryMock.go`)

Hand-written function-field mock, identical shape to `userRepositoryMock.go`:
- Import the port aliased (`pkg<Domain> "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/<domain>"`) and add the assertion `var _ pkg<Domain>.<Entity>Repository = &<Entity>RepositoryMock{}`.
- `<Entity>RepositoryMock{ <Method>Fn func(ctx context.Context, request pkg<Domain>.<Method>Request) (*pkg<Domain>.<Method>Response, error) }` — one `<Method>Fn` field per method.
- Each method delegates straight to its `Fn` field (`return m.<Method>Fn(ctx, request)`). No mutex/counters and no `t *testing.T` parameter unless a test needs call-count assertions.

This mock is the testing seam for any future service/adaptor. Per the `go-unit-tests` skill the Mongo queries themselves (filters, pagination, regex) are **DB-backed and out of scope** for pure-Go tests — they belong to `tests/integration` (live server on :8080); do not write a DB test here.

## Finish

1. `go build ./... && go vet ./... && go test ./<domain-paths>/...` — must all pass. (Integration tests under `tests/integration` need a live server on :8080 and will fail when one isn't running; that's unrelated.)
2. Tell the user exactly what was created and restate the boundary: **the entity now has a repository but is NOT reachable over REST** — exposing it still needs an `<Entity>Service` port, a `<Entity>ServiceRESTAdaptor`, and wiring in `cmd/app/serviceProvider.go` + `setupAPIServer.go` (use the `scaffold-service` skill).