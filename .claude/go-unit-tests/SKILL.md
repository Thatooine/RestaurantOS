---
name: go-unit-tests
description: Write, run, and maintain meaningful pure-Go unit tests for this project. Use when adding tests for a service, adaptor, or validation, or when checking coverage. Covers the two pure-Go test tiers (pure logic and mock-first logic/adaptor), the hand-written mock convention, the mandatory happy-path-AND-error-case rule, the table-driven validation style, and the run/coverage commands. No database or running app required.
---

# Go unit tests

This skill is for **pure-Go unit tests only** — tests that run with nothing but `go test`. There is **no expectation that the app, MongoDB, Redis, or any external service is running.** Every test here passes against mocked collaborators and in-memory logic; if a test would need a live Mongo/Redis/HTTP service, it does not belong to this skill (mock the collaborator instead, or leave that invariant to the separate integration suite at `tests/integration`, which drives a live server on `:8080`).

Module is `github.com/bash/the-dancing-pony-v2-rnyfbr`. Two conventions are non-negotiable in this repo:

1. **Every method is tested for the happy-day path AND each error branch.** A test that only proves the success case is incomplete — the error mapping, the fail-closed guard, the not-found path are where bugs hide.
2. **Mock as much as possible.** Test a unit against mocked collaborators (the `pkg` interfaces), never a live database. Owner pinning, error mapping, defaulting, branching — all of it is provable with hand-written mocks and zero infrastructure.

## The two pure-Go test tiers

| Tier | Where | Backing | Use for |
|------|-------|---------|---------|
| **Pure logic** | `pkg/<domain>/*_test.go` (in-package `package <domain>`) | none | `Validate()`, parsing, codecs — table-driven |
| **Logic / adaptor (mock-first)** | `internal/pkg/<domain>/*_test.go` (in-package) | **mocks** | services, orchestration — branch coverage against mocked repos/collaborators |

Both run offline with plain `go test ./...` — no `docker compose`, no DSN, no server process. Default to the mock-first tier for anything with branches; use the pure-logic tier for anything that is just data-in/data-out.

## Mock-first unit tests (the primary technique)

Mocks are **hand-written function-field stubs**, one per interface, living next to the implementation in `internal/pkg/<domain>/`. We do not use gomock/mockery — plain `go` tooling, no codegen. Canonical examples: `internal/pkg/users/userRepositoryMock.go` and `internal/pkg/authentication/accessTokenCreatorServiceMock.go`.

### Mock structure

```go
import pkgUsers "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"

// var _ asserts the mock satisfies the interface — drift fails the build.
var _ pkgUsers.UserRepository = &UserRepositoryMock{}

type UserRepositoryMock struct {
	// one function field per interface method, named <Method>Fn
	CreateUserFn func(ctx context.Context, request pkgUsers.CreateUserRequest) (*pkgUsers.CreateUserResponse, error)
	GetUserFn    func(ctx context.Context, request pkgUsers.GetUserRequest) (*pkgUsers.GetUserResponse, error)
	// ...one per method...
}

func (m *UserRepositoryMock) GetUser(ctx context.Context, request pkgUsers.GetUserRequest) (*pkgUsers.GetUserResponse, error) {
	return m.GetUserFn(ctx, request)
}
```

Each method delegates straight to its `<Method>Fn` field. Keep the field signature **identical to the interface method** — no extra `t *testing.T` or `m` parameters (the closure already captures `t` from the test). A test wires only the `Fn` fields it exercises; calling an unset field panics, which is fine because a correct test sets exactly the ones the unit should reach. Add a `sync.Mutex` + a `<Method>Invocations int` counter back **only** when a test needs call-count assertions or drives the unit concurrently.

### Writing the test — mock as stub AND spy

Put the test **in-package** (`package <domain>`, same package as the impl and the mock under `internal/pkg/<domain>`) — see `userRegistrationServiceImpl_test.go`. The mock lives in that same package, so there is no import cycle and no need for an external `_test` package. Construct the impl by setting only the collaborator fields the test needs (e.g. `&UserRegistrationServiceImpl{userRepository: repo}`).

- **Happy path** — the unit returns a value; assert it mapped/returned it correctly, and assert *what the unit passed the collaborator* (e.g. that the service pinned `UserID` to the claim, or passed the right email). The mock is a spy here — assert inside the `Fn`, fail with `t.Fatalf`.
- **Error path** — the collaborator's `Fn` returns `errs.ErrNotFound` (etc.); assert the unit surfaces / maps it correctly and that `errors.Is` still matches the sentinel through any `%w` wrap.
- **Fail-closed / not-called path** — put `t.Fatal(...)` *inside* a collaborator `Fn` to prove that collaborator is **never called** (e.g. `CreateUser` must not run when the user already exists; an unauthorized request must not reach the repo).

For REST adaptors, drive the handler with `httptest.NewRequest` + `httptest.NewRecorder`. When the handler reads the login claim, inject it onto the request context the same way the auth middleware does (via the `authentication` context helper) so the claim-pinning and fail-closed-on-missing-claim branches are both covered.

## Pure-logic tests — table-driven `Validate()`

The established style (`pkg/<domain>/<x>Validations_test.go`, in-package): a `valid()` constructor, a `mutate func(*Request)`, and a `wantErr bool`. The happy-path-and-errors rule is satisfied structurally — the first row is valid, every other row breaks exactly one field.

```go
func validReq() CreateUserRequest { return CreateUserRequest{Name: "Frodo", Email: "frodo@shire.com"} }

tests := []struct {
	name    string
	mutate  func(*CreateUserRequest)
	wantErr bool
}{
	{"valid", func(r *CreateUserRequest) {}, false},
	{"missing name", func(r *CreateUserRequest) { r.Name = "" }, true},
	{"missing email", func(r *CreateUserRequest) { r.Email = "" }, true},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		req := validReq()
		tt.mutate(&req)
		err := req.Validate()
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
		}
	})
}
```

## Out of scope: invariants that ARE the database

Some behaviour can only be proven against a real engine — mocking it would test the mock, not the system. **These are explicitly out of scope for this skill** and belong to the integration suite at `tests/integration/api_test.go`, which drives a live server on `:8080` (infrastructure this skill assumes is absent):

- **Mongo query correctness** — that a `bson.M` filter, pagination (`offset`/`limit`), or a `$regex` search actually selects the right documents.
- **Unique-index enforcement** — a uniquely-indexed field rejecting a duplicate insert.
- **End-to-end auth + routing** — middleware, the login claim, rate limiting, and the route wiring composing correctly.

When you reach one of these in a unit test, **stop and mock the boundary instead**: assert that the unit *calls the repository with the right scoped/filtered request*, not that the Mongo query filters. If a behaviour genuinely cannot be tested without a live Mongo/HTTP stack, leave a `// integration: see tests/integration` note rather than pulling infrastructure into this tier.

## Running tests & coverage

No setup, no services — these run on a clean checkout:

```bash
go test ./...                                   # all pure-Go tests; no DB needed
go test ./internal/pkg/users/ -run TestIssueToken -v   # one package / one test (regex)
go test -cover ./...                            # per-package coverage %

# Coverage report, then inspect uncovered branches in the browser
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

Chase coverage of **branches that carry behaviour** (each error return, each gate), not a percentage. Generated/trivial getters are not worth a line.

## Maintenance

- Mark helpers with `t.Helper()` so failures point at the call site.
- Keep tests deterministic: never call `time.Now()`/`rand` and assert against a literal — inject the value or assert structurally (`!IsZero()`). With no database or clock in play, a pure-Go test that flakes is a test bug, not infrastructure.
- When you change an interface, the mock's `var _ Interface = ...` assertion breaks the build — update the mock in the same change.
- Scaffolding a repository or service? Pair this skill with `scaffold-crud-repository` / `scaffold-service`, which call back here for the table-driven `Validate()` test and the mock-first impl test.
