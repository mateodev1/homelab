# HomeLab

A personal homelab management system built as a Go + React monorepo.

Repository: https://github.com/mateodev1/homelab.git

## Architecture

```
frontend/ (React + TypeScript + Vite)
    └─ /api/* proxy ──→ backend/ (Go HTTP API)
                              └─ shared/  (pure domain types)
                              └─ data/homelab.db (SQLite)

cli/ (Go CLI + MCP server)
    ├─ cmd/homelab      — human CLI
    ├─ cmd/homelab-mcp  — Model Context Protocol server (stdio)
    └─ shared/          — pure domain types
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

- Go 1.24+ (CLI/MCP; backend still builds with 1.23+)
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

## CLI Configuration

The `homelab` CLI (under `cli/`) persists its configuration to a JSON file so you
don't have to repeat flags:

```
~/.config/homelab/config.json
```

The directory honors `HOMELAB_CONFIG_DIR` for ad-hoc overrides; the file is
written with permission `0600` (owner read/write only; best-effort on Windows).
The file contains only `base_url`, `api_key`, and `env` (snake_case JSON);
`require_auth` is never persisted — it is always derived as `env == "production"`.

### Resolution precedence

Configuration is resolved, highest to lowest:

1. Explicit flags (`--base-url`, `--api-key`, `--env`)
2. `HOMELAB_BASE_URL` / `HOMELAB_API_KEY` / `HOMELAB_ENV` environment variables
   (an empty env value is treated as unset)
3. The config file fields
4. Built-in defaults (`base_url` = `http://localhost:8080`, `env` = development)

### `homelab login`

Persist the config file with `homelab login` (it never contacts the backend):

```bash
# Development (default): stores env=development and an empty API key, no prompt.
homelab login

# Production: reads the API key from stdin (no echo on a TTY; one line when piped).
homelab login --env production
echo "$HOMELAB_API_KEY" | homelab login --env production
```

Notes:

- `--env` defaults to `development`; when omitted, the existing config file's
  `env` is reused; an explicit `--env=production` overrides the file.
- `--base-url` is honored into the persisted file when provided.
- `--api-key` is ignored by `login` (the key comes from stdin in production) and
  emits a stderr warning if passed.
- On an empty production input, `login` exits non-zero and writes no file.

## MCP server

`homelab-mcp` exposes the backend as Model Context Protocol tools over stdio.
It reuses the same config file and HTTP client as the CLI, so the AI never sees
the API key. Run `homelab login` once; the MCP process only reads config.

### Build

```bash
go build -o homelab-mcp ./cli/cmd/homelab-mcp
# or: task build  (then install the binary somewhere on your PATH)
```

### Tools

| Tool | Backend |
|------|---------|
| `health` | `GET /api/health` |
| `todo_list` / `todo_get` / `todo_create` / `todo_update` / `todo_done` / `todo_delete` | `/api/todos` |
| `project_list` / `project_get` / `project_create` / `project_update` / `project_delete` | `/api/projects` |
| `secret_product_*`, `secret_project_*`, `secret_environment_list` | `/api/products/...` |
| `secret_list` / `secret_create` / `secret_reveal` / `secret_update` / `secret_delete` / `secret_export` | `/api/products/.../secrets` |

`todo_update` uses boolean `clear_due_date`, `clear_issue_type`, and
`clear_project_id` to null nullable fields (do not set a value and its `clear_*`
flag together). There is no `login` tool — that stays interactive on the CLI.

### OpenCode (this project only)

Project-scoped config lives in the repo root `opencode.json` (merged over your
global `~/.config/opencode/opencode.json` when you open this workspace). It is
**not** registered globally.

```json
{
  "mcp": {
    "homelab": {
      "type": "local",
      "command": ["go", "run", "./cli/cmd/homelab-mcp"],
      "enabled": true
    }
  }
}
```

Restart opencode after pulling this config. Optional faster startup:

```bash
go build -o bin/homelab-mcp ./cli/cmd/homelab-mcp
# then set command to ["./bin/homelab-mcp"] in opencode.json
```

Auth mirrors the backend: development is open; production requires
`homelab login --env production` (or `HOMELAB_API_KEY`).

## Secret Manager

The authenticated frontend exposes a secret manager at `/secrets` with the
hierarchy `Product → Project → Environment → Secret`. New secret projects get
`development`, `staging`, and `production` environments automatically.

Secret values are encrypted at rest with `SECRETS_ENCRYPTION_KEY`. Existing
deployments without that variable use a stable compatibility key derived from
`API_KEY`; configure the dedicated variable before rotating `API_KEY` so stored
secrets remain decryptable.

The environment screen supports revealing individual values, copying the full
environment, and downloading the complete environment as `.env`. Bulk export
always includes every secret in the selected environment, regardless of UI
filters.

## Deployment

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
