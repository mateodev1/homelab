---
name: homelab-api-contract
description: "Trigger: backend endpoint, todo API, issue API, auth0, authorization, frontend API client changes. Document and keep homeLab's Go backend contract and frontend client in sync."
license: MIT
metadata:
  author: mateo
  version: "1.0"
---

## Activation Contract

Use this skill when changing:
- Any file under `backend/internal/handler/`, `backend/internal/service/`, `backend/internal/domain/`, `backend/internal/store/`
- The todo/issue wire format (fields, statuses, priorities, validation)
- Anything related to authentication/authorization for the API
- `frontend/src/api/todos.ts` or `frontend/src/types/todo.ts`

## Hard Rules

- **No backend authorization exists today.** Auth0 (`@auth0/auth0-react`) only gates the frontend route `/_authenticated` (UI-level redirect). The frontend never attaches an `Authorization` header; `frontend/src/api/todos.ts` calls `fetch('/api/todos')` with no auth headers. The Go backend (`backend/internal/handler/middleware.go`) has no auth/JWT middleware — only CORS, logging, and panic recovery. Treat `/api/todos*` as unauthenticated until this changes. Do not assume Auth0 protects the API.
- Read `references/api-contract.md` before changing endpoint shapes, statuses, priorities, or validation rules.
- When you change the backend contract (fields, statuses, priorities, validation, routes), update in the SAME change/PR:
  1. `references/api-contract.md` (source of truth for this skill)
  2. `frontend/src/types/todo.ts` and `frontend/src/api/todos.ts`
  3. Go tests in `backend/internal/handler/*_test.go` and `backend/internal/service/*_test.go`
  4. A SQLite migration in `backend/internal/store/migrations.go` if columns change
  5. Frontend tests in `frontend/src/api/todos.test.ts`
- Follow existing layering strictly: `store` (SQLite) → `service` (business rules, `TodoPatch`) → `handler` (`net/http`, JSON in/out). Do not collapse layers.
- Routing uses stdlib `http.ServeMux` with two patterns per resource: `/api/todos` (collection) and `/api/todos/` (item, ID parsed from path suffix). Follow this pattern for new resources — do not introduce a router library.

## Decision Gates

| Situation | Action |
| --- | --- |
| Adding a new todo/issue field | Add migration → domain struct → service patch → handler request/response → frontend type/client → tests → update `references/api-contract.md` |
| Adding real API authorization | Add middleware in `backend/internal/handler/middleware.go`, wire it in `backend/cmd/api/main.go`'s middleware chain, and update the "No backend authorization exists today" rule above plus `references/api-contract.md` |
| Changing valid statuses/priorities | Update `domain.ValidStatuses` (or priority bounds in `handler/todo.go`), migration if needed, frontend types, and `references/api-contract.md` |
| Unsure if a change affects the contract | If it touches request/response JSON shape, validation, routes, or auth, it affects the contract — update this skill |

## Execution Steps

1. Read `references/api-contract.md` for current endpoints, request/response shapes, and validation rules.
2. Make the backend change following the store → service → handler layering.
3. Update or add Go tests per the layer touched (unit + integration, per `CONTRIBUTING.md` checklist).
4. Update `frontend/src/types/todo.ts` and `frontend/src/api/todos.ts` to match, plus `frontend/src/api/todos.test.ts`.
5. Update `references/api-contract.md` to reflect the new contract.
6. Run `task lint` and `task test` before pushing (per `CONTRIBUTING.md`).

## Output Contract

Return:
- Which layers were touched (store/service/handler/frontend).
- Whether `references/api-contract.md` was updated.
- Whether a migration was added.
- Whether the "no backend auth" assumption still holds, or whether auth was added (and where).

## References

- `references/api-contract.md` — endpoint list, request/response JSON shapes, status/priority enums, and auth model detail.
- `CONTRIBUTING.md` (repo root) — lint/test commands and coverage requirements.
