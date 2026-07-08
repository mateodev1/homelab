# homeLab API Contract Reference

Source of truth for the Go backend HTTP API and its authorization model.
Verified against code on 2026-07-08 (`backend/internal/handler/todo.go`,
`backend/internal/domain/todo.go`, `backend/internal/handler/middleware.go`,
`backend/cmd/api/main.go`, `frontend/src/routes/__root.tsx`,
`frontend/src/routes/_authenticated.tsx`, `frontend/src/api/todos.ts`,
`docker/nginx.conf.template`, `docker/frontend.Dockerfile`).

## Entrypoint / routing

- `backend/cmd/api/main.go` wires: `store.New(db)` → `service.NewTodoService(store)` → `handler.NewTodoHandler(svc)`.
- Routing uses stdlib `http.ServeMux` — no third-party router.
- Middleware chain (outer to inner): `RecoveryMiddleware(LoggingMiddleware(CORSMiddleware(protected)))`, where `protected` is `handler.AuthMiddleware(apiKeyForAuth, validator, mux)` — always wired, regardless of `ENV`.
- `CORSMiddleware` sets `Access-Control-Allow-Origin: *` and allows `GET, POST, PUT, DELETE, OPTIONS`. No auth-related headers or checks.
- Persistence: SQLite via `modernc.org/sqlite` (pure-Go driver, registered with blank import in `main.go`). DB path from `DB_PATH` env var (default `/data/homelab.db`).

## Authorization model — IMPORTANT

**The Go backend always enforces authorization; two credential types are accepted in parallel.**
`backend/internal/handler/middleware.go` defines `AuthMiddleware(apiKey, validator, next)`:
- Requires header `Authorization: Bearer <token>` on every request.
- Exempts `OPTIONS` (CORS preflight) and `GET /api/health` — these never require a credential.
- Accepts EITHER an exact match against `apiKey` (M2M client credential, see below) OR a JWT validated by `validator` (Auth0 access token: RS256 signature verified against JWKS, `iss`, `aud`, `exp`).
- Returns `401 {"error": "unauthorized"}` if neither matches.

Wiring (`backend/cmd/api/main.go`):
- `AUTH0_DOMAIN` and `AUTH0_AUDIENCE` are read from the environment and are **always required** (dev and prod) — the process `log.Fatalf`s at startup if either is missing. `handler.NewJWTValidator` builds a `keyfunc.Keyfunc` against `https://${AUTH0_DOMAIN}/.well-known/jwks.json` (background refresh/rotation via `github.com/MicahParks/keyfunc/v3` + `github.com/golang-jwt/jwt/v5`).
- `API_KEY` is read from the environment too. If `ENV=production` and it's empty, the process fails fast (unchanged from before). The `apiKey` value passed into `AuthMiddleware` is only non-empty when `ENV=production` — in any other environment the API_KEY path is effectively disabled, only JWTs work.
- `AuthMiddleware` is **always** added to the chain, regardless of `ENV` (no more "open in dev" mode — the JWT path is always live).

Auth0 (`@auth0/auth0-react`) is used **only** in the frontend for login, but the backend now
validates the resulting access tokens directly:
- `frontend/src/routes/__root.tsx` wraps the app in `<Auth0Provider>` with `VITE_AUTH0_DOMAIN` / `VITE_AUTH0_CLIENT_ID`. **`audience` is not yet configured in `authorizationParams`** — until it is, `getAccessTokenSilently()` won't return an API-scoped JWT that the backend's `aud` check will accept.
- `frontend/src/routes/_authenticated.tsx` is a route-level gate: if `!isAuthenticated`, it renders a login prompt instead of `<Outlet />`. This blocks UI navigation only.
- `frontend/src/api/todos.ts` calls `fetch('/api/todos', ...)` with **no `Authorization` header** and no access token retrieval. Combined with the two points above, the frontend cannot successfully call the API yet — this is tracked as separate, not-yet-implemented frontend work (add `audience` to `Auth0Provider`, attach `Authorization: Bearer <token>` in `todos.ts`).

Production frontend URL: `https://todos.matdev.site` (per Auth0 SPA redirect config).

## Data model

`backend/internal/domain/todo.go`:

```go
type Todo struct {
    ID        int64
    Title     string
    Body      string
    Status    string   // see ValidStatuses
    Priority  int      // 0-3, see priority note below
    DueDate   *string  // nullable, string format (no time.Time)
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Status enum (`domain.ValidStatuses`)

- `todo`
- `in_progress`
- `done`
- `cancelled`

### Priority

Priority is an **integer 0–3** (not a string enum). Validated in the handler:
`0 <= priority <= 3` on both create and update. There is no named
`low|medium|high|urgent` enum in code — memory referencing string priority
levels is inaccurate; verify against `backend/internal/handler/todo.go` if this
changes.

### Migration history (`backend/internal/store/migrations.go`)

- Legacy `done` boolean column exists; migration added `status`, `priority`, `due_date` columns and backfilled `status = 'done'` where `done = 1`.
- New fields: `status TEXT NOT NULL DEFAULT 'todo'`, `priority INTEGER NOT NULL DEFAULT 0`, `due_date TEXT NULL`.

## Endpoints

Base path: `/api/todos`

| Method | Path | Handler | Notes |
| --- | --- | --- | --- |
| GET | `/api/todos` | `ListTodos` | Returns JSON array, `[]` when empty |
| POST | `/api/todos` | `CreateTodo` | Body: `{title, body, priority, due_date}` |
| GET | `/api/todos/{id}` | `GetTodo` | 404 via `{"error": "not found"}` if missing |
| PUT | `/api/todos/{id}` | `UpdateTodo` | Body: `{title?, body?, status?, priority?, due_date?}` — partial patch |
| DELETE | `/api/todos/{id}` | `DeleteTodo` | 204 No Content on success |

Health check: `GET /api/health` via `handler.NewHealthHandler` (not detailed here — see `backend/internal/handler/health.go`).

### Request/response shapes

**POST /api/todos** request:
```json
{ "title": "string (required)", "body": "string", "priority": 0, "due_date": "string|null" }
```
Validation: `title` required (non-empty), `priority` in `[0,3]`.

**PUT /api/todos/{id}** request (all fields optional, partial patch via `service.TodoPatch`):
```json
{ "title": "string?", "body": "string?", "status": "string?", "priority": "int?", "due_date": "string|null?" }
```
Validation: if `title` present it must be non-blank (trimmed); if `status` present it must be one of `ValidStatuses`; if `priority` present it must be in `[0,3]`.
Note: `due_date` uses `**string` (pointer-to-pointer) to distinguish "not provided" from "explicitly set to null".

**Response shape** (`todoResponse` in `backend/internal/handler/todo.go`), same for all endpoints returning a Todo:
```json
{
  "id": 1,
  "title": "string",
  "body": "string",
  "status": "todo|in_progress|done|cancelled",
  "priority": 0,
  "due_date": "string|null",
  "created_at": "RFC3339 timestamp",
  "updated_at": "RFC3339 timestamp"
}
```

Errors: `{"error": "message"}` with appropriate HTTP status (400, 404, 405, 500).

## Frontend client

`frontend/src/api/todos.ts` — plain `fetch` wrapper, no auth headers, calls relative paths (`/api/todos`, etc.):
- `getTodos(signal?)` → `GET /api/todos`
- `createTodo(payload)` → `POST /api/todos`
- `getTodoById(id)` → `GET /api/todos/{id}`
- `updateTodo(id, payload)` → `PUT /api/todos/{id}`
- `deleteTodo(id)` → `DELETE /api/todos/{id}`, expects 204

**How the SPA reaches the protected API**: `docker/nginx.conf.template` reverse-proxies
`/api/` to the backend transparently (`proxy_pass http://backend:8080;` plus `Host`/
`X-Real-IP` only) — it does **not** inject any `Authorization` header. Whatever the
browser sends is what the backend sees. Since `frontend/src/api/todos.ts` sends no
`Authorization` header today, and `Auth0Provider` has no `audience` configured, the SPA
currently gets `401` from every `/api/todos*` call against this backend (dev and prod
alike, per the always-on `AuthMiddleware`) until the frontend-side work described above
is done. In dev, `frontend/vite.config.ts`'s own proxy (`/api` → `http://backend:8080`)
is used instead of nginx, but the same `401` applies since the backend enforces auth
regardless of `ENV`.

Types live in `frontend/src/types/todo.ts` (`Todo`, `CreateTodoPayload`, `UpdateTodoPayload`, `ApiError`).

## Testing conventions (see CONTRIBUTING.md)

- Go: unit tests per layer (`domain`, `store`, `service`, `handler`) plus `handler` integration tests (`TestIntegration_*`) and middleware tests.
- Frontend: `frontend/src/api/todos.test.ts` mocks global `fetch` and asserts exact call args/payloads.
- CI enforces ≥60% coverage across `./backend/... ./shared/...`.
