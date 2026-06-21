# HomeLab

A personal homelab management system built as a Go + React monorepo.

## Architecture

```
frontend/ (React + TypeScript + Vite)
    └─ /api/* proxy ──→ backend/ (Go HTTP API)
                              └─ shared/  (pure domain types)
                              └─ data/homelab.db (SQLite)

cli/ (Go CLI)
    └─ shared/  (pure domain types)
```

The Go code follows hexagonal architecture:

```
domain → store → service → handler
```

- **domain**: pure types and repository interfaces — no I/O
- **store**: SQLite implementation of domain interfaces
- **service**: business logic, depends on domain interfaces only
- **handler**: HTTP layer, depends on service interfaces only

## Prerequisites

- Go 1.23+
- Node 20+
- pnpm 9+
- Task (taskfile.dev)
- Docker + Docker Compose

## Quick Start

```bash
cp .env.example .env
task dev
```

## Development

```bash
# Start backend + frontend dev servers via Docker Compose
task dev

# Run all tests
task test

# Run Go tests only
task test:go

# Run frontend tests only
task test:frontend
```

## Testing

```bash
# All tests (Go race detector + Vitest)
task test

# Go only
task test:go

# Frontend only
task test:frontend
```

## Linting

```bash
# All linters
task lint

# Go only (golangci-lint)
task lint:go

# Frontend only (Biome)
task lint:frontend
```

## Deployment

```bash
# Build Docker images
task docker:build

# Start all services
docker compose up
```
