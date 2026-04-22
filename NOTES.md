# Implementation Notes

Submission for the Entain (Neds) backend technical test. Five tasks from the
brief delivered as five stacked PRs, each targeting the previous, with a Phase 0
chore commit landing dev tooling and hygiene fixes on `main` before any feature
branch diverged.

## Architecture

```
            HTTP                   gRPC
  client ─► api (:8000) ────────► racing (:9000) ──► racing/db/racing.db
                │       ────────► sports (:9001) ──► sports/db/sports.db
                │
          grpc-gateway (reverse proxy)
```

Three independent Go modules. Each service owns its own `go.mod`, port, DB, and
gRPC server. The `api` gateway does no business logic — it translates HTTP/JSON
to gRPC via `google.api.http` annotations and forwards to the right backend.

## Design decisions (and the trade-offs)

### Derived status, not persisted
The brief needs `status` ∈ {`OPEN`, `CLOSED`} based on `advertised_start_time`
vs `now`. Three realistic designs:

- Persist + cron sync: needs background infra, goes stale between runs.
- Persist + compute-on-write: wrong the instant `advertised_start_time` passes `now`.
- **Derive at read time**: always correct, no infra, one `time.Before` per row.

Read-time derivation is the only option that can't be stale. Cost is negligible
at this scale. `OPEN = 0` lines up with the proto3 zero-value so an unset field
defaults to the common case.

### Input validation lives in the service layer, not the repo
`sort_by` is user-supplied and goes into an `ORDER BY` clause that can't be
parameter-bound (SQL drivers can only bind values, not identifiers). An
allowlist is unavoidable.

I put the allowlist in the service. Rationale: the repo should be a thin SQL
builder; it shouldn't know about gRPC status codes. Rejecting in the service
with `codes.InvalidArgument` maps cleanly to HTTP 400 at the gateway. An
adversarial test (`TestListRaces_RejectsInvalidSortBy`) proves both the status
code and that the repo is never called on bad input.

A `ListRacesOptions` struct crosses the service → repo boundary so the repo
receives pre-validated, transport-agnostic input.

### Typed sentinel for not-found, not collapsed error mapping
`GetRace` has one of the most common "looks-right-until-prod-breaks" bugs:
mapping *every* repo error to `codes.NotFound`. A DB outage then reports as
HTTP 404, which ops can't act on.

The repo exports `var ErrRaceNotFound = errors.New("race not found")`. The
service uses `errors.Is` (not `==`, so it works through `%w` wrapping) to
discriminate: `ErrRaceNotFound` → `codes.NotFound`, anything else →
`codes.Internal`. Unit tests cover both branches; a bufconn integration test
proves the status code round-trips through real gRPC.

### Context threaded to the repo
`RacesRepo.List/Get` and `EventsRepo.List` take `ctx context.Context` as the
first param; SQL calls use `QueryContext`. Cancellations and deadlines from the
gRPC handler propagate all the way to SQLite. `defer rows.Close()` and
`rows.Err()` checks are present on every query path (both of which the template
was missing).

### Sports scope: intentionally minimal
Brief: *"...implements a similar API to racing. We'll leave it up to you to
determine what you might think a sports event is made up off, but it should at
minimum have an id, a name and an advertised_start_time."*

I kept sports to `ListEvents` + `visible_only` filter — matching the scope the
brief explicitly allows ("at minimum"). Sort, status, and `GetEvent` are not
implemented; the patterns are all in racing and porting them across is ~30 LOC
each. I chose scope restraint over feature parity to keep the submission
focused and reviewable.

### PR workflow: five stacked feature PRs + Phase 0 chore commits on main
HR's brief: *"5x PRs in total. Each PR should target the previous."* Exactly
delivered. Feature branches: `feature/ET-001__visible-filter` →
`feature/ET-002__sort` → `feature/ET-003__status` →
`feature/ET-004__get-race` → `feature/ET-005__sports`.

Before PR 1 diverged from `main`, three chore commits landed: Phase 0 (Go 1.22
bump, Makefile, CI rewrite, deprecated-import swap, ctx threading, rows
hygiene), a grpc/protobuf version bump with regenerated stubs, and testify
addition plus test helpers. This kept each feature PR a pure feature diff —
reviewers opening PR 1 see the visible_only feature, not a grpc bump buried
inside it.

## Testing strategy

- **Repo tests** use real in-memory SQLite (`sql.Open("sqlite3", ":memory:")`) —
  real driver, real SQL execution. A mock would hide schema/SQL bugs.
- **Service tests** use hand-rolled mock repos (no framework) to test
  proto→options translation, error classification, and validation rejection.
- **One integration test** (`racing/service/integration_test.go`) boots a real
  gRPC server over `bufconn` and dials it as a client — proves `codes.NotFound`
  survives the wire. Targeted where regression risk is highest: error
  classification.
- **Convention**: `require.NoError` for preconditions (halt cleanly on setup
  failure), `assert` for the actual check (record failure, continue).

## Production-readiness gaps (what I'd do next)

Short list, ordered roughly by leverage:

1. **Graceful shutdown.** `grpcServer.GracefulStop()` on SIGINT/SIGTERM in all
   three services. ~15 LOC each.
2. **Pagination** on `ListRaces` / `ListEvents`. Cursor-based
   (`after_id` + `limit`) scales; offset-based doesn't.
3. **Structured logging** (`zap` or `zerolog`) with request IDs propagated via
   gRPC metadata so requests are traceable across services.
4. **Metrics**: `grpc_prometheus` interceptor for RED metrics on every RPC;
   latency histograms on `GetRace`.
5. **OpenAPI export** from the annotated protos (`protoc-gen-openapiv2`) so
   consumers get a spec alongside the service.
6. **Single proto source-of-truth.** The gateway and each service currently
   hold near-duplicate `.proto` files; consolidating to one package consumed
   by both modules would eliminate drift risk.
7. **Docker + compose.** Dockerfile per service; compose file for local dev
   and CI parity.
8. **Sort tiebreakers.** `ORDER BY advertised_start_time ASC` has no tiebreak
   today; add `, id ASC` for deterministic ordering under pagination.
9. **NOT NULL constraints + migration tooling.** Hand-written `seed()` works
   for a demo; `goose`/`golang-migrate` is the production answer.

## On using AI

I used Claude and Codex as pairing tools — independent code reviews of a
reference submission (to calibrate what senior reviewers flag), and to
accelerate mechanical work (Makefile boilerplate, test scaffolds, proto regen).
Design calls and architectural trade-offs are mine. Every claim in the reviews
has file:line anchors so cross-checking was straightforward.
