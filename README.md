# HomeLab

A personal homelab management system built as a Go + React monorepo.

Repository: https://github.com/mateodev1/homelab.git

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
- Node 24+
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

Deploys are triggered by pushing a semver tag. CI builds the `prod` targets for
`backend` and `frontend`, pushes them to GHCR, then a self-hosted runner pulls
and restarts the stack on the homelab server.

```bash
# Cut a release (triggers build-and-push + deploy jobs in .github/workflows/ci.yml)
git tag vX.Y.Z
git push origin vX.Y.Z
```

What the `deploy` job runs on the server (`~/homelab`):

```bash
docker login ghcr.io -u "$GHCR_USER" --password-stdin
docker compose pull
docker compose up -d
docker image prune -f
```

Local/manual equivalents:

```bash
# Build Docker images locally
task docker:build

# Start all services
docker compose up
```

## Installing the API Client Skill Standalone

To use the `homelab-api-client` skill in another environment without cloning the
whole monorepo:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/mateodev1/homelab.git
cd homelab
git sparse-checkout set skills/homelab-api-client
```

This pulls only `skills/homelab-api-client/` — no backend, frontend, or other
project files.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the full pre-push checklist: lint commands, test inventory, coverage requirements, and CI pipeline details.
