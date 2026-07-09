# homeLab Todo API — Endpoint Reference

## Connection

- Base URL: `API_BASE_URL` env var (e.g. `https://homelab.example.com`).
- Auth: `API_KEY` env var, sent as `Authorization: Bearer <API_KEY>` on every request.
  - **Status: enforced in production.** Verified against the live deployment: requests without the header (or with a wrong value) get `401 {"error": "unauthorized"}`.
- Content type: `application/json` for all request/response bodies.

## Enums

- `priority`: integer, `0` to `3` (inclusive). No named levels — just the integer.
- `status`: string, one of `todo`, `in_progress`, `done`, `cancelled`.

## Endpoints

### `GET /api/health`

Health check. No auth required (send the header anyway per contract above).

Response `200`:
```json
{
  "status": "ok",
  "db_ok": true
}
```

### `GET /api/todos`

List all todos.

Response `200`:
```json
[
  {
    "id": 1,
    "title": "Buy milk",
    "body": "2% or whole",
    "status": "todo",
    "priority": 2,
    "due_date": "2026-07-10",
    "created_at": "2026-07-08T12:00:00Z",
    "updated_at": "2026-07-08T12:00:00Z"
  }
]
```

### `POST /api/todos`

Create a todo.

Request:
```json
{
  "title": "Buy milk",
  "body": "2% or whole",
  "priority": 2,
  "due_date": "2026-07-10"
}
```

Rules:
- `title` is required, non-empty.
- `priority` must be `0-3`.
- `due_date` is optional, nullable string.

Response `201`: same shape as a list item (see above), with server-assigned `id`, `status: "todo"`, `created_at`, `updated_at`.

Error `400`:
```json
{ "error": "title is required" }
```
```json
{ "error": "priority must be between 0 and 3" }
```

### `GET /api/todos/{id}`

Fetch one todo by numeric ID.

Response `200`: same shape as a list item.

Error `400` (non-numeric id):
```json
{ "error": "invalid id" }
```

Error `404`:
```json
{ "error": "not found" }
```

### `PUT /api/todos/{id}`

Partial update — all fields optional, only send what changes.

Request:
```json
{
  "title": "Buy oat milk",
  "body": "unsweetened",
  "status": "in_progress",
  "priority": 1,
  "due_date": "2026-07-12"
}
```

Rules:
- `title`, if present, must be non-empty after trimming.
- `status`, if present, must be one of the valid enum values.
- `priority`, if present, must be `0-3`.
- `due_date` may be set to `null` to clear it.

Response `200`: full updated todo (same shape as list item).

Error `400`:
```json
{ "error": "status must be one of: todo, in_progress, done, cancelled" }
```

Error `404`:
```json
{ "error": "not found" }
```

### `DELETE /api/todos/{id}`

Delete a todo by numeric ID.

Response: `204` No Content, empty body.

Error `404`:
```json
{ "error": "not found" }
```

## Common error shape

All error responses share this shape:
```json
{ "error": "human-readable message" }
```
