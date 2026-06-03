# Uptime Monitor

A small HTTP service in Go where a user registers URLs and has them checked on a
schedule by a background worker. Each check records the HTTP status code and
response time, and the API exposes per-monitor history and uptime statistics.

I built this as my first backend project written from scratch — no starter
template and no web framework. The goal was to understand how the parts of a REST
API actually fit together, so I stayed on Go's standard library and only reached
for a third-party package where writing it myself would have been a mistake (the
PostgreSQL driver and password hashing).

## Tech stack

- Go — standard library for HTTP, routing, DB access, logging, and the JWT
  signing (`net/http`, `database/sql`, `context`, `log/slog`, `crypto/hmac`)
- PostgreSQL
- `pgx` as the database driver, used through `database/sql`
- `golang.org/x/crypto/bcrypt` for password hashing
- Docker and Docker Compose

## How it works

Two things run inside one process:

1. **HTTP API** — registration, login, and CRUD for monitors, plus history and
   stats endpoints. Every route that touches user data sits behind JWT auth
   middleware.
2. **Background worker** — a goroutine driven by a `time.Ticker`. On each tick it
   asks the database which monitors are due for a check, pings them concurrently,
   and writes the result (status code + response time) into the `checks` table.

Request flow for a protected endpoint:

```
request
  -> auth middleware            (validate Bearer token, put user id in context)
  -> handler                    (decode JSON, read user id, validate path)
  -> storage (repository)       (parameterized SQL query)
  -> JSON response
```

## Design notes

These are choices I made deliberately while building it:

- **Standard-library routing.** Go 1.22's `http.ServeMux` matches on method and
  path patterns (`GET /monitor/{id}`, read with `r.PathValue("id")`), so a router
  library wasn't necessary for this scope.
- **Hand-written JWT.** I implemented HS256 myself — base64url header and payload,
  signed with HMAC-SHA256, verified with a constant-time compare (`hmac.Equal`) —
  to learn how a token is actually structured and checked. For real production use
  I would switch to a vetted library such as `golang-jwt`.
- **Repository pattern.** All SQL lives in the `storage` package behind methods
  like `InsertMonitor` and `GetMonitorByID`; handlers never contain SQL.
- **Context timeouts.** Every database call runs under a 3s `context.WithTimeout`,
  and the worker's outbound HTTP client has a 10s timeout, so a slow target or a
  stuck query can't hang the process.
- **Per-user isolation.** Ownership is enforced in the query itself
  (`WHERE id = $1 AND user_id = $2`), so one user can't read or change another
  user's monitors even with a guessed id.
- **"Due for check" decided in SQL.** The worker selects monitors where
  `last_checked_at IS NULL OR last_checked_at + (interval '1 second' *
  check_interval) <= NOW()`, instead of tracking timers per monitor in memory.
- **Graceful shutdown.** `signal.NotifyContext` catches SIGINT/SIGTERM. The HTTP
  server drains via `srv.Shutdown`, the worker's context is cancelled, and `main`
  waits on a `sync.WaitGroup` so in-flight checks finish before exit.
- **Request hardening.** Request bodies are capped with `http.MaxBytesReader`, and
  the JSON decoder uses `DisallowUnknownFields` to reject unexpected input.
- **Structured logging** via `log/slog` (JSON handler).

## Data model

**users**

| column     | type         | notes                  |
|------------|--------------|------------------------|
| id         | serial PK    |                        |
| username   | varchar(255) |                        |
| password   | varchar(60)  | bcrypt hash            |
| email      | varchar(255) | unique index on lower(email) |
| created_at | timestamp    |                        |
| updated_at | timestamp    |                        |

**monitors**

| column          | type         | notes                                    |
|-----------------|--------------|------------------------------------------|
| id              | serial PK    |                                          |
| user_id         | int FK       | references users(id), on delete cascade  |
| url             | varchar(255) |                                          |
| check_interval  | int          | seconds between checks                   |
| last_checked_at | timestamp    | nullable; indexed                        |
| created_at      | timestamp    |                                          |
| updated_at      | timestamp    |                                          |

**checks**

| column        | type        | notes                                   |
|---------------|-------------|-----------------------------------------|
| id            | bigserial PK|                                         |
| monitor_id    | int FK      | references monitors(id), on delete cascade; indexed |
| status_code   | int         | HTTP status from the ping (0 on error)  |
| response_time | int         | milliseconds                            |
| created_at    | timestamp   |                                         |
| updated_at    | timestamp   |                                         |

## API

All responses are JSON with the shape `{ "status", "message", "data" }`
(`data` is omitted when empty).

| Method | Path                    | Auth   | Description                          |
|--------|-------------------------|--------|--------------------------------------|
| GET    | /health                 | public | liveness check                       |
| POST   | /users/register         | public | create an account                    |
| POST   | /users/login            | public | log in, returns a JWT                |
| POST   | /monitor                | bearer | create a monitor                     |
| GET    | /monitor                | bearer | list the caller's monitors           |
| GET    | /monitor/{id}           | bearer | get one monitor                      |
| PATCH  | /monitor/{id}           | bearer | update url / check_interval          |
| DELETE | /monitor/{id}           | bearer | delete a monitor                     |
| GET    | /monitor/{id}/checks    | bearer | last 50 checks, newest first         |
| GET    | /monitor/{id}/stats     | bearer | total checks, avg response, uptime % |

Protected routes expect an `Authorization: Bearer <token>` header.

### Examples

Register, then log in to get a token:

```bash
curl -X POST localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"sam","email":"sam@example.com","password":"secret123"}'

curl -X POST localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"sam@example.com","password":"secret123"}'
# -> {"status":"success","message":"Login Success","data":{"token":"<jwt>"}}
```

Create a monitor and read its stats:

```bash
curl -X POST localhost:8080/monitor \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","check_interval":30}'

curl localhost:8080/monitor/1/stats -H "Authorization: Bearer <jwt>"
# -> {"status":"success","message":"get monitor stats",
#     "data":{"total_checks":12,"avg_response_time":143,"uptime_percentage":100}}
```

## Running locally

Requirements: Docker (and Go 1.22+ if you want to run the server outside a
container).

Configuration comes from environment variables (loaded from a `.env` file if
present):

| variable     | example                                                            |
|--------------|--------------------------------------------------------------------|
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/uptime_monitor?sslmode=disable |
| JWT_SECRET   | a long random string                                               |
| PORT         | 8080                                                               |

**With Docker Compose**

```bash
docker compose up --build
```

This starts PostgreSQL and the API. The schema in `migrations/001_init.sql` is
applied once against the database (see the migrations note below), after which
the API is available on `http://localhost:8080`.

**Applying the schema**

```bash
# against a running database
psql "$DATABASE_URL" -f migrations/001_init.sql
```

## Limitations and next steps

This is a learning project, and there are things I left out on purpose or would
do next:

- **No automated tests yet.** The handlers and the storage layer are the obvious
  places to start (`net/http/httptest` + a test database).
- **Input validation is minimal** — e.g. URL format and `check_interval` bounds
  aren't checked yet.
- **JWT is hand-rolled** for learning; a production version would use a maintained
  library and rotate the signing secret.
- **Migrations are applied manually.** A migration tool, or auto-running the SQL
  on database init, would make setup one step.
- **Checks history is capped at the latest 50** with no pagination.
- **No TLS / alerting** — in a real deployment this would sit behind a reverse
  proxy, and a "monitor went down" notification would be the next feature.
