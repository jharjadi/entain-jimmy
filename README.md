## Entain BE Technical Test

This test has been designed to demonstrate your ability and understanding of technologies commonly used at Entain.

Please treat the services provided as if they would live in a real-world environment.

### Directory Structure

- `api`: REST gateway that forwards JSON requests onto the backing gRPC services.
- `racing`: Racing service (gRPC on `:9000`, SQLite backend).
- `sports`: Sports service (gRPC on `:9001`, SQLite backend).

```
entain/
├─ api/
│  ├─ proto/
│  ├─ main.go
├─ racing/
│  ├─ db/
│  ├─ proto/
│  ├─ service/
│  ├─ main.go
├─ sports/
│  ├─ db/
│  ├─ proto/
│  ├─ service/
│  ├─ main.go
├─ Makefile
├─ NOTES.md
├─ README.md
```

### Getting Started

1. Install Go (1.23+).

```bash
brew install go
```

... or [see here](https://golang.org/doc/install).

2. Install `protoc`.

```bash
brew install protobuf
```

... or [see here](https://grpc.io/docs/protoc-installation/).

3. Install the proto generation tools (one-time).

```bash
make install-tools
```

4. Start all three services.

```bash
make run
```

This backgrounds `racing` on `:9000`, `sports` on `:9001`, and the `api` gateway on `:8000`.
Use `make kill` to stop them.

### Example requests

**List races** (default order: `advertised_start_time ASC`):

```bash
curl -s -X POST http://localhost:8000/v1/list-races \
  -H 'Content-Type: application/json' \
  -d '{"filter":{}}' | jq
```

**Visible-only + custom sort:**

```bash
curl -s -X POST http://localhost:8000/v1/list-races \
  -H 'Content-Type: application/json' \
  -d '{"filter":{"visible_only":true,"sort_by":"id","sort_direction":"DESC"}}' | jq
```

**Invalid `sort_by` returns HTTP 400** (allowlist-validated in the service layer):

```bash
curl -si -X POST http://localhost:8000/v1/list-races \
  -H 'Content-Type: application/json' \
  -d '{"filter":{"sort_by":"DROP TABLE races"}}' | head -1
# HTTP/1.1 400 Bad Request
```

**Get a single race** (404 for genuinely missing ids, 500 for infrastructure errors):

```bash
curl -s http://localhost:8000/v1/races/1 | jq
curl -si http://localhost:8000/v1/races/99999 | head -1
# HTTP/1.1 404 Not Found
```

**List sports events:**

```bash
curl -s -X POST http://localhost:8000/v1/sports/events \
  -H 'Content-Type: application/json' \
  -d '{"filter":{"visible_only":true}}' | jq
```

### Development

| Command          | Purpose                                                      |
|------------------|--------------------------------------------------------------|
| `make build`     | Build all three binaries                                     |
| `make run`       | Background all three services                                |
| `make kill`      | Kill whatever is bound to `:8000` / `:9000` / `:9001`        |
| `make test`      | `go test ./... -race -cover` across all three modules        |
| `make lint`      | `go vet ./...` across all three modules                      |
| `make proto`     | Regenerate all `.pb.go` / `.pb.gw.go` stubs                  |

### Tasks (per HR brief)

1. **Visible-only filter** on `ListRaces` — `visible_only` bool on the filter; non-breaking.
2. **Default ordering** by `advertised_start_time` with caller-configurable `sort_by` / `sort_direction`. Allowlist-validated; invalid values return HTTP 400.
3. **Derived `status`** on `Race` (`OPEN` / `CLOSED`), computed at read time from `advertised_start_time` vs `now`.
4. **`GetRace` RPC** with proper error classification (`NotFound` for missing ids, `Internal` for everything else — avoids the "every error is 404" trap).
5. **Separate `sports` service** with its own module, port, DB, and gRPC server; gateway routes it at `POST /v1/sports/events`.

See [NOTES.md](NOTES.md) for design decisions, trade-offs, and production-readiness gaps.

### Good Reading

- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Google API Design](https://cloud.google.com/apis/design)
- [Go Modules](https://golang.org/ref/mod)
- [Ubers Go Style Guide](https://github.com/uber-go/guide/blob/2910ce2e11d0e0cba2cece2c60ae45e3a984ffe5/style.md)
