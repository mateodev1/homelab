# Contributing — Pre-Push Checklist

All checks below must pass before pushing or opening a PR. CI runs exactly these same commands.

---

## Quick run (all at once)

```bash
task lint   # lint Go + frontend
task test   # test Go + frontend
```

---

## Go

### Lint

```bash
# Via Task
task lint:go

# Direct (requires golangci-lint v2.3.0+)
cd backend && golangci-lint run ./...

# Via Docker (no local Go/golangci-lint needed)
docker run --rm \
  -v $(pwd)/backend:/app -w /app \
  golangci/golangci-lint:v2.3.0 \
  golangci-lint run ./...
```

Config: `backend/.golangci.yml` (version `"2"`, enabled linters: `errcheck`, `govet`, `staticcheck`, `unused`)

### Tests

```bash
# Via Task
task test:go

# Direct
cd backend && go test -race ./...
```

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

> `package.json` declares `pnpm.onlyBuiltDependencies` to allow `@biomejs/biome` and `esbuild`
> to run their postinstall scripts (native binary download). If those are blocked, biome will
> fall back to a slow JS shim and may incorrectly lint `node_modules`.

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
| `build` | always | `go build ./backend/cmd/api`, `go build ./cli/cmd/homelab` | — |
| `build-and-push-backend` | tag only | `docker/build-push-action@v6`, pushes to GHCR | `docker/backend.Dockerfile` |
| `build-and-push-frontend` | tag only | `docker/build-push-action@v6`, pushes to GHCR | `docker/frontend.Dockerfile` |
| `deploy` | tag only | `docker compose pull && up -d` on self-hosted runner | `environment: homelab` |

---

## Deploy to production

```bash
git tag v0.1.7 && git push origin v0.1.7
```

This triggers the full CI pipeline. If all tests pass, the Docker images are built, pushed to GHCR, and deployed automatically.
