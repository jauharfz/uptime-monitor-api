# Uptime Monitor
![CI](https://github.com/jauharfz/uptime-monitor-api/actions/workflows/ci.yml/badge.svg)

A small HTTP service in Go where a user registers URLs and has them checked on a
schedule by a background worker. Each check records the HTTP status code and
response time, the API exposes per-monitor history and uptime statistics, and an
optional webhook (e.g. Discord) fires when a site goes down or recovers.

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

1. **REST API** — registration, login, and CRUD for monitors, plus history and
   stats endpoints. Every route that touches user data sits behind JWT auth
   middleware.
2. **Background worker** — a goroutine driven by a `time.Ticker`. On each tick it
   asks the database which monitors are due for a check, pings them concurrently,
   and writes the result (status code + response time) into the `checks` table.
   When a monitor flips between up and down, the worker POSTs a notification to
   that monitor's `webhook_url` if one is set — pointing it at a Discord webhook
   gives you down/up alerts in a channel.

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
  path patterns (`GET /monitors/{id}`, read with `r.PathValue("id")`), so a router
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
- **Alerts only on state transitions.** The webhook fires when a monitor's
  up/down state flips, not on every failed check — a site that is down for an
  hour produces one alert, not one per check.
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
| id               | serial PK    |                                          |
| user_id          | int FK       | references users(id), on delete cascade  |
| url              | varchar(255) |                                          |
| check_interval   | int          | seconds between checks (min 10)          |
| last_checked_at  | timestamp    | nullable; indexed                        |
| webhook_url      | varchar(255) | optional alert target; empty = no alerts |
| last_status_code | int          | status from the previous check, used to detect up/down transitions |
| created_at       | timestamp    |                                          |
| updated_at       | timestamp    |                                          |

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

All success responses are JSON with the shape `{ "status", "message", "data" }`
(`data` is omitted when empty). The full contract lives in
[openapi.yaml](openapi.yaml) — browse it rendered at
[/docs](https://uptime-monitor-api.duckdns.org/docs) on the live instance, or
import it into Postman or Insomnia to try the endpoints.

| Method | Path                           | Auth   | Description                          |
|--------|--------------------------------|--------|--------------------------------------|
| GET    | /health                        | public | liveness check                       |
| GET    | /docs                          | public | interactive API docs (Scalar)        |
| GET    | /openapi.yaml                  | public | the OpenAPI spec itself              |
| POST   | /api/v1/users/register         | public | create an account                    |
| POST   | /api/v1/users/login            | public | log in, returns a JWT                |
| POST   | /api/v1/monitors               | bearer | create a monitor                     |
| GET    | /api/v1/monitors               | bearer | list the caller's monitors           |
| GET    | /api/v1/monitors/{id}          | bearer | get one monitor                      |
| PATCH  | /api/v1/monitors/{id}          | bearer | update url / check_interval / webhook_url |
| DELETE | /api/v1/monitors/{id}          | bearer | delete a monitor                     |
| GET    | /api/v1/monitors/{id}/checks   | bearer | last 50 checks, newest first         |
| GET    | /api/v1/monitors/{id}/stats    | bearer | total checks, avg response, uptime % |

Protected routes expect an `Authorization: Bearer <token>` header.

### Examples

Register, then log in to get a token:

```bash
curl -X POST localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"sam","email":"sam@example.com","password":"secret123"}'

curl -X POST localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"sam@example.com","password":"secret123"}'
# -> {"status":"success","message":"Login Success","data":{"token":"<jwt>"}}
```

Create a monitor (optionally with a webhook for down/up alerts) and read its
stats:

```bash
curl -X POST localhost:8080/api/v1/monitors \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","check_interval":30,
       "webhook_url":"https://discord.com/api/webhooks/<id>/<token>"}'

curl localhost:8080/api/v1/monitors/1/stats -H "Authorization: Bearer <jwt>"
# -> {"status":"success","message":"get monitor stats",
#     "data":{"total_checks":12,"avg_response_time":143,"uptime_percentage":100}}
```

## Tests

The handlers and storage layer are tested against a real PostgreSQL database
rather than mocks, so the SQL runs too. `TestMain` (in
`internal/api/setup_test.go`) connects to a separate `uptime_monitor_test`
database, drops the schema and reapplies every file in `migrations/` in order
before the run, and starts the background worker so the whole process is
exercised. Each
handler test drives an endpoint with `net/http/httptest` and checks the status
code and JSON body.

```bash
# expects a throwaway Postgres reachable at DATABASE_URL
go test ./...
```

## Running locally

Requirements: Docker (and Go 1.26+ if you want to run the server outside a
container).

Configuration comes from environment variables, loaded from a `.env` file (copy
`.env.example` to start). Docker Compose reads the Postgres credentials and builds
the API's `DATABASE_URL` from them, so `DATABASE_URL` itself is only needed when
you run the binary directly.

| variable          | used by          | example                                  |
|-------------------|------------------|------------------------------------------|
| POSTGRES_USER     | compose          | postgres                                 |
| POSTGRES_PASSWORD | compose          | a strong random string                   |
| POSTGRES_DB       | compose          | uptime_monitor                           |
| JWT_SECRET        | api              | a long random string                     |
| PORT              | api              | 8080                                     |
| DATABASE_URL      | api (direct run) | postgres://postgres:…@localhost:5454/uptime_monitor?sslmode=disable |

**With Docker Compose**

```bash
cp .env.example .env   # then fill in real values
docker compose up --build
```

This starts PostgreSQL and the API. The `migrations/` folder is mounted into the
Postgres init directory, so on a fresh database volume every numbered migration
is applied in order automatically. The API is then available on
`http://localhost:8080` (Postgres is published on `localhost:5454`).

**Applying the schema manually** — needed if you run Postgres yourself outside
Compose, or when a new migration lands on an already-initialised volume (the init
directory only runs on first boot):

```bash
for f in migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done
```

## Deployment

A live instance runs on an Azure Ubuntu VM. Docker Compose builds the API and
PostgreSQL on the host, Caddy sits in front as a reverse proxy and terminates TLS
with an automatically renewed Let's Encrypt certificate, and the hostname is a
DuckDNS subdomain. The containers' published ports are bound to `localhost`, so
only Caddy reaches them from outside.

Live: [`/health`](https://uptime-monitor-api.duckdns.org/health) ·
[`/docs`](https://uptime-monitor-api.duckdns.org/docs)

## Limitations and next steps

This is a learning project, and there are things I left out on purpose or would
do next:

- **Validation is light.** URLs with no scheme get `https://` prepended and
  `check_interval` has a 10-second floor, but URLs aren't otherwise validated —
  including against private/internal addresses (SSRF), which a public deployment
  would need to reject.
- **JWT is hand-rolled** for learning; a production version would use a maintained
  library and rotate the signing secret.
- **Migrations are plain numbered SQL files** applied in order; a real migration
  tool (e.g. goose) would track which ones have already run instead of relying on
  fresh-volume init or manual psql.
- **Checks history is capped at the latest 50** with no pagination.
- **Webhook delivery is fire-and-forget.** One POST per up/down transition with a
  3s timeout — no retries or backoff if the webhook endpoint happens to be down.
