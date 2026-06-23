# Plan — Fix CI warnings

## Tasks

- [ ] Bump action versions in `ci.yml` to node24-compatible releases
- [ ] Add `cache-dependency-path: backend/go.sum` to `test-go` and `build` jobs

## Detail

### Problem 1 — Node.js 20 deprecation (5 warnings)

GitHub runners now run on Node.js 24. Actions that internally target `node20` generate
a deprecation warning. The fix is to upgrade to the major version where each action
updated their `runs.using` field to `node24`.

| Current | Fix | Evidence |
|---------|-----|---------|
| `actions/checkout@v4` | `actions/checkout@v5` | v4.x still node20; v5+ released for node24 |
| `actions/setup-node@v4` | `actions/setup-node@v6` | Latest is v6.4.0; v6 migrated to node24 |
| `pnpm/action-setup@v4` | `pnpm/action-setup@v6` | Latest is v6.0.9 |
| `actions/setup-go@v5` | `actions/setup-go@v6` | v6.2.0 changelog: "Update Node.js version in action.yml" |

Apply to ALL jobs in `ci.yml` (lint-go, test-go, lint-frontend, test-frontend, build).

### Problem 2 — `go.sum` cache path missing (2 warnings)

`setup-go` with `cache: true` looks for `go.sum` at the repo root.
This repo has it at `backend/go.sum`. The `lint-go` job already has the correct
`cache-dependency-path: backend/go.sum` — the other two jobs are missing it.

Jobs to fix:
- `test-go` (line ~34): add `cache-dependency-path: backend/go.sum`
- `build` (line ~99): add `cache-dependency-path: backend/go.sum`

### File

`.github/workflows/ci.yml` — only file that needs changes.
`release.yml` uses no Go or Node setup actions → no changes needed.
