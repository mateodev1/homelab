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
- Middleware chain (outer to inner): `RecoveryMiddleware(LoggingMiddleware(CORSMiddleware(protected)))`, where `protected` is `APIKeyMiddleware(apiKey, mux)` when `ENV=production`, or just `mux` otherwise.
- `CORSMiddleware` sets `Access-Control-Allow-Origin: *` and allows `GET, POST, PUT, DELETE, OPTIONS`. No auth-related headers or checks.
- Persistence: SQLite via `modernc.org/sqlite` (pure-Go driver, registered with blank import in `main.go`). DB path from `DB_PATH` env var (default `/data/homelab.db`).

## Authorization model — IMPORTANT

**The Go backend enforces a single, global API key in production.**
`backend/internal/handler/middleware.go` defines `APIKeyMiddleware(apiKey, next)`:
- Requires header `Authorization: Bearer <API_KEY>` on every request.
- Exempts `OPTIONS` (CORS preflight) and `GET /api/health` — these never require the key.
- Returns `401 {"error": "unauthorized"}` on missing/incorrect key.

Wiring (`backend/cmd/api/main.go`):
- `API_KEY` is read from the environment (populated from the root `.env` via `env_file` in `docker-compose.yml` / Task's `dotenv`).
- If `ENV=production` and `API_KEY` is empty, the process **fails fast** (`log.Fatalf`) at startup.
- The middleware is only added to the chain when `ENV=production`. In development (`ENV=development` or unset), all `/api/*` routes are open, matching the pre-existing behavior.

Auth0 (`@auth0/auth0-react`) is used **only** in the frontend, and is a separate, unrelated
mechanism from the API key:
- `frontend/src/routes/__root.tsx` wraps the app in `<Auth0Provider>` with `VITE_AUTH0_DOMAIN` / `VITE_AUTH0_CLIENT_ID`.
- `frontend/src/routes/_authenticated.tsx` is a route-level gate: if `!isAuthenticated`, it renders a login prompt instead of `<Outlet />`. This blocks UI navigation only.
- `frontend/src/api/todos.ts` calls `fetch('/api/todos', ...)` with **no `Authorization` header** and no access token retrieval (`getAccessTokenSilently` is not used). This means the frontend does **not** currently send the API key — it will get `401`s if pointed at a production backend until it's updated to send `Authorization: Bearer <API_KEY>`.

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

**How the SPA reaches the protected API in production**: the frontend JS never sends the
`Authorization` header itself (the API key is never baked into the JS bundle). Instead,
in prod (`docker/frontend.Dockerfile`, stage `prod`) nginx serves the static build and
reverse-proxies `/api/` to the backend, injecting the header server-side:
- `docker/nginx.conf.template` uses `nginx:alpine`'s built-in envsubst-on-templates
  entrypoint (`/etc/nginx/templates/*.template` → `/etc/nginx/conf.d/*`) to render
  `proxy_set_header Authorization "Bearer ${API_KEY}";` from the container's `API_KEY`
  env var at container startup (not at build time).
- The `frontend` container in the production deployment must receive `API_KEY` (same
  value as the backend's) as a runtime env var. This is configured outside this repo, in
  the compose file used by the self-hosted deploy runner.
- In dev, `frontend/vite.config.ts`'s own proxy (`/api` → `http://backend:8080`) is used
  instead, and the backend doesn't enforce the key when `ENV=development`, so no header
  is needed.

Types live in `frontend/src/types/todo.ts` (`Todo`, `CreateTodoPayload`, `UpdateTodoPayload`, `ApiError`).

## Testing conventions (see CONTRIBUTING.md)

- Go: unit tests per layer (`domain`, `store`, `service`, `handler`) plus `handler` integration tests (`TestIntegration_*`) and middleware tests.
- Frontend: `frontend/src/api/todos.test.ts` mocks global `fetch` and asserts exact call args/payloads.
- CI enforces ≥60% coverage across `./backend/... ./shared/...`.
