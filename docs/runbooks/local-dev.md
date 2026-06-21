# Local Development Runbook

Step-by-step guide to get the homelab project running locally.

## Prerequisites

Install the following tools before cloning:

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | https://go.dev/dl/ |
| Node | 20+ | https://nodejs.org or `nvm` |
| pnpm | 9+ | `npm install -g pnpm` |
| Task | latest | https://taskfile.dev/installation/ |
| Docker | latest | https://docs.docker.com/get-docker/ |
| Docker Compose | v2+ | Bundled with Docker Desktop |

Verify installations:

```bash
go version      # go1.23.x or later
node --version  # v20.x or later
pnpm --version  # 9.x or later
task --version  # Task x.x.x or later
docker version  # Engine: 25.x or later
```

## Clone and Configure

```bash
git clone https://github.com/mateo/homelab.git
cd homelab

# Copy env template and edit as needed
cp .env.example .env
```

Default `.env` values work out of the box for local development.

## Install Dependencies

```bash
# Go workspace — downloads no external deps for the scaffold
go work sync

# Frontend dependencies
pnpm --dir frontend install
```

## Start Dev Servers

```bash
# Starts backend + frontend via Docker Compose (hot-reload enabled)
task dev
```

Services will be available at:
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/api/health

## Run Tests

```bash
# All tests (Go race detector + Vitest)
task test

# Go only
task test:go

# Frontend only
task test:frontend
```

## Run Linters

```bash
# All linters
task lint

# Go only (golangci-lint)
task lint:go

# Frontend only (Biome)
task lint:frontend
```

## Build Binaries

```bash
task build
```

Output binaries: `bin/api` (backend), `bin/homelab` (CLI).

## Docker Compose (Full Stack)

```bash
# Build and start all services
docker compose up --build

# Stop all services
docker compose down

# Rebuild images after dependency changes
task docker:build
```

## Database

SQLite database is stored at `./data/homelab.db` (gitignored).
The `data/` directory is bind-mounted into the backend container.

To reset the database:

```bash
rm data/homelab.db
task dev   # re-creates on startup
```

## Troubleshooting

### Port already in use

```bash
# Find and kill the process using port 8080
lsof -ti:8080 | xargs kill -9
```

### Go workspace out of sync

```bash
go work sync
```

### Frontend node_modules out of date

```bash
pnpm --dir frontend install --frozen-lockfile
```

### Docker build cache stale

```bash
docker compose build --no-cache
```
