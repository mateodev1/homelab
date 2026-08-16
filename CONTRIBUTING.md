# Contributing — Pre-Push and Deploy Checklist

Use this checklist before pushing, opening a PR, or creating a production tag.

The local commands below are the source of truth for deploy preparation. CI runs the same
categories of checks, but some jobs use GitHub Actions directly instead of the `task` wrapper.

---

## Quick run (all at once)

```bash
pnpm --dir frontend install
task lint   # lint Go + frontend
task test   # test Go + frontend
task build  # build Go backend + CLI + MCP
pnpm --dir frontend run type-check
pnpm --dir frontend run coverage
rm -f api homelab homelab-mcp
git status --short
```

Expected result before deploy:

- Every command exits with code `0`.
- `git status --short` shows only intentional source/doc changes.
- No root-level `api`, `homelab`, or `homelab-mcp` binaries remain after `task build`.

---

## Go

### Lint

```bash
# Via Task
task lint:go

# Direct (requires golangci-lint v2.3.0+)
cd backend && golangci-lint run ./...
cd cli && golangci-lint run --config ../backend/.golangci.yml ./...
cd shared && golangci-lint run --config ../backend/.golangci.yml ./...

# Via Docker (no local Go/golangci-lint needed)
docker run --rm \
  -v $(pwd)/backend:/app -w /app \
  golangci/golangci-lint:v2.3.0 \
  golangci-lint run ./...
```

Config: `backend/.golangci.yml` (version `"2"`, enabled linters: `errcheck`, `govet`, `staticcheck`, `unused`)

This repo uses `go.work` with separate Go modules. Do not run `golangci-lint run ./...`
from the repo root.

### Tests

```bash
# Via Task
task test:go

# Direct
cd backend && go test -race ./...
cd cli && go test -race ./...
cd shared && go test -race ./...
```

This repo uses `go.work` with separate Go modules. Do not run `go test -race ./...`
from the repo root.

#### Test coverage per package

| Package | Tests |
|---|---|
| `internal/domain` | `TestTodoZeroValue`, `TestTodoFieldAssignment`, `TestErrNotFound`, `TestHealthStatus` |
| `internal/store` | `TestCreate_Insert`, `TestCreate_MultipleTodos`, `TestGetAll_ReturnsAllInserted`, `TestGetAll_EmptyDB`, `TestGetByID_Found`, `TestGetByID_NotFound`, `TestUpdate_NotFound`, `TestUpdate_ChangesTitle`, `TestDelete_Removes`, `TestDelete_NotFound` |
| `internal/service` | `TestCreateTodo_AssignsID`, `TestCreateTodo_PropagatesStoreError`, `TestListTodos`, `TestListTodos_EmptyStore`, `TestGetTodo_Found`, `TestGetTodo_NotFound`, `TestUpdateTodo_PersistsChanges`, `TestUpdateTodo_NotFound`, `TestDeleteTodo_Removes`, `TestDeleteTodo_NotFound` |
| `internal/handler` (unit) | `TestListTodos_OK`, `TestListTodos_EmptyReturnsArray`, `TestListTodos_DoneFilter_True`, `TestListTodos_DoneFilter_InvalidParam`, `TestCreateTodo_Created`, `TestCreateTodo_BadJSON`, `TestGetTodo_Found`, `TestGetTodo_NotFound`, `TestGetTodo_DBError`, `TestUpdateTodo_OK`, `TestUpdateTodo_BlankTitle`, `TestUpdateTodo_NotFound`, `TestDeleteTodo_NoContent`, `TestDeleteTodo_NotFound`, `TestGetHealth_OK` |
| `internal/handler` (integration) | `TestIntegration_CRUDCycle`, `TestIntegration_GetNotFound`, `TestIntegration_DoneFilter`, `TestIntegration_PUTBlankTitle`, `TestIntegration_CORSPreflight` |
| `internal/handler` (middleware) | `TestCORSMiddleware_SetsHeaders`, `TestCORSMiddleware_Options`, `TestRecoveryMiddleware_Panic`, `TestLoggingMiddleware_Logs` |

CI enforces ≥ 60% coverage across `./backend/... ./shared/...`.

---

## Frontend

### Install

```bash
pnpm --dir frontend install
```

> `frontend/pnpm-workspace.yaml` must include `packages: ['.']` and allow `@biomejs/biome`
> and `esbuild` to run postinstall scripts. Without that, CI can fail during install or miss
> the native Biome/Rollup binaries needed for lint/test execution.

### Lint & format (Biome)

```bash
# Via Task
task lint:frontend

# Direct
pnpm --dir frontend run lint

# Auto-fix formatter issues
pnpm --dir frontend run lint:fix
```

Config: `frontend/biome.json`
- Formatter: 2-space indent, 100-char line width, single quotes in JS/TS
- `node_modules` and `dist` are excluded via `files.ignore`

### Tests (Vitest)

```bash
# Via Task
task test:frontend

# Direct
pnpm --dir frontend test --run

# With coverage report
pnpm --dir frontend run coverage
```

#### Test coverage per file

| File | Tests |
|---|---|
| `src/api/todos.test.ts` | `getTodos` success/error, `createTodo` success/error, `getTodoById` success/error, `updateTodo` success/error, `deleteTodo` 204/error |
| `src/hooks/useTodos.test.ts` | loads on mount, error on load failure, `addTodo` appends, `toggleTodo` updates done, `removeTodo` removes, action errors captured |
| `src/components/TodoForm.test.tsx` | submit disabled on empty, controlled input, calls `onAdd` and clears on submit |
| `src/components/TodoItem.test.tsx` | renders title/checkbox, `onToggle` on click, `onDelete` on click, line-through when done |
| `src/components/TodoList.test.tsx` | loading state, error state, empty state, renders items |
| `src/components/NoteCard.test.tsx` | render, pin, done toggle, delete, title edit/revert/escape/enter, body edit/escape, color picker, done class |
| `src/components/NoteForm.test.tsx` | placeholder render, expand, commit with/without title, escape collapse, color picker, click outside |
| `src/components/NoteGrid.test.tsx` | loading, error, empty, only-pinned, only-unpinned, mixed sections |
| `src/App.test.tsx` | renders NoteForm + NoteGrid from hook state, loading spinner, search filtering, clear search, error state |

---

## CI pipeline

The GitHub Actions workflow (`.github/workflows/ci.yml`) runs these jobs in order:

```
lint-go ──────────────────────────────────┐
                                           ▼
lint-frontend ──→ test-frontend ──→ build (Go backend + CLI)
                                           ▲
lint-go ──→ test-go ──────────────────────┘
```

On tag push (`v*.*.*`), after all CI jobs pass:

```
build ──→ build-and-push-backend ──┐
          build-and-push-frontend ──┴──→ deploy
```

| Job | Trigger | Tool | Config |
|---|---|---|---|
| `lint-go` | always | `golangci-lint-action@v9`, golangci-lint `v2.3.0` | `backend/.golangci.yml` |
| `test-go` | always | `go test -race`, coverage ≥ 60% | `backend/go.mod` |
| `lint-frontend` | always | Biome `check` + `tsc --noEmit` | `frontend/biome.json` |
| `test-frontend` | always | Vitest with coverage | `frontend/vitest.config` |
| `build` | always | `go build ./backend/cmd/api`, `go build ./cli/cmd/homelab`, `go build ./cli/cmd/homelab-mcp` | — |
| `build-and-push-backend` | tag only | `docker/build-push-action@v6`, pushes to GHCR | `docker/backend.Dockerfile` |
| `build-and-push-frontend` | tag only | `docker/build-push-action@v6`, pushes to GHCR | `docker/frontend.Dockerfile` |
| `deploy` | tag only | `docker compose pull && up -d` on self-hosted runner | `environment: homelab` |

---

## Deploy to production

Production deploys are triggered by pushing a semver tag. Do not tag until the deploy
preparation checklist is green locally and `main` contains the release commit.

### 1. Prepare locally

```bash
pnpm --dir frontend install
task lint
task test
task build
pnpm --dir frontend run type-check
pnpm --dir frontend run coverage
rm -f api homelab homelab-mcp
git status --short
```

If any command fails, fix the failure and rerun the full checklist. Do not deploy from a
partially verified state.

### 2. Push main

```bash
git push origin main
```

Wait for the `main` CI run to pass before tagging when possible.

### 3. Create and push the production tag

```bash
git tag v0.1.18
git push origin v0.1.18
```

This triggers the full CI pipeline. If all tests pass, the Docker images are built, pushed to GHCR, and deployed automatically.

### 4. Watch CI and verify production

```bash
gh run list --limit 5
gh run watch <run-id> --exit-status
curl https://todos.matdev.site/api/health
```

Expected production health response:

```json
{"status":"ok","db_ok":true}
```

If the tag workflow fails before deploy, fix the issue on `main` and create a new patch tag.
Do not reuse or force-move a production tag.
