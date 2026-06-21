# Architecture

## Overview

HomeLab is a monorepo consisting of three Go modules and one React frontend, orchestrated by a Taskfile and deployed via Docker Compose.

## Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Frontend | React 18 + TypeScript + Vite | Web UI served on :5173 |
| Backend API | Go 1.23, net/http | REST API served on :8080 |
| CLI | Go 1.23 (cobra) | Command-line management tool |
| Database | SQLite (file at `data/homelab.db`) | Zero-infra persistence |
| Reverse proxy | Vite dev proxy (`/api/*` → `:8080`) | Local dev; Nginx in prod |
| Container runtime | Docker Compose | Local dev and CI build |
| CI/CD | GitHub Actions | Lint, test, build, release |
| Remote access | Tailscale | Secure homelab VPN tunnel |
| Frontend linter | Biome | Single-tool lint + format |
| Go linter | golangci-lint + depguard | Static analysis + import guard |

## Module Structure

```
/
├── go.work                         # Go workspace — ties three modules together
├── backend/                        # module: github.com/mateo/homelab/backend
│   ├── go.mod
│   └── cmd/api/main.go             # wiring only — no business logic
│   └── internal/
│       ├── domain/                 # pure types + repository interfaces
│       ├── store/                  # SQLite implementation
│       ├── service/                # business logic
│       └── handler/                # HTTP handlers
│
├── cli/                            # module: github.com/mateo/homelab/cli
│   ├── go.mod
│   └── cmd/homelab/main.go
│
├── shared/                         # module: github.com/mateo/homelab/shared
│   ├── go.mod
│   └── pkg/domain/types.go         # shared pure domain types (no I/O)
│
└── frontend/                       # React + TypeScript + Vite
    ├── src/
    ├── biome.json
    └── vite.config.ts
```

## Hexagonal Architecture (Backend)

The backend follows hexagonal (ports & adapters) architecture. Dependencies always point inward:

```
domain  ←  store  ←  service  ←  handler  ←  cmd/api/main.go
  │                                              (wiring only)
  └── pure types + interfaces; zero I/O
```

### Layer Contracts

```go
// domain/ — pure types, stdlib only
type Todo struct { ID int64; Title string; Done bool; CreatedAt time.Time }
type Repository interface {
    List(ctx context.Context) ([]Todo, error)
    Create(ctx context.Context, t Todo) (Todo, error)
}

// store/ — implements Repository; imports domain + database/sql
type SQLiteStore struct { db *sql.DB }

// service/ — business logic; depends on Repository interface (injected)
type TodoService struct { repo domain.Repository }

// handler/ — HTTP; depends on service interface (injected)
type TodoHandler struct { svc TodoServicer }

// cmd/api/main.go — wiring only
store := store.New(db)
svc   := service.New(store)
h     := handler.New(svc)
```

### shared/ Purity Rule

`shared/` is enforced clean by `depguard` in `.golangci.yml`. It MUST NOT import:
- `database/sql`
- `os`
- `io`
- `net/http`

This is validated automatically on every `golangci-lint run ./...`.

## Data Flow

```
Browser :5173
  └─ Vite dev proxy /api/* ──→ backend :8080
                                   └─ SQLiteStore
                                         └─ ./data/homelab.db
```

## CI/CD Pipeline

```
PR / push to main
  ├─ lint-go       (golangci-lint run ./...)
  ├─ test-go       (go test -race -coverprofile; coverage ≥ 60%)
  ├─ lint-frontend (pnpm biome check)
  └─ test-frontend (pnpm vitest run)
        ↓ all pass
  build job        (go build ./backend/cmd/api ./cli/cmd/homelab)

tag v*.*.*
  └─ release.yml → docker build → ghcr.io (semver + SHA tags)
```

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Module isolation | 3 separate `go.mod` + `go.work` | Each binary ships independently; workspace keeps cross-module type safety |
| shared/ purity guard | depguard in golangci.yml | Automated enforcement; catches I/O imports before merge |
| DB engine | SQLite in `data/` | Zero infra for homelab; no DB container complexity |
| Frontend linter | Biome (single tool) | One config, faster, no plugin version skew |
| Non-root Docker UID | UID 1000 | Matches typical Linux dev UID; simple, portable |
| Env handling | `.env` + `.env.example` | Homelab scope; no external secrets service needed |
