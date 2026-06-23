# Plan — Consolidate CI and release into one workflow

## Tasks

- [ ] Add `push: tags: v*.*.*` trigger to `ci.yml`
- [ ] Add `build-and-push-backend`, `build-and-push-frontend`, and `deploy` jobs to `ci.yml` (only run on tag pushes, depend on CI jobs passing)
- [ ] Delete `release.yml`

## Detail

### Goal

Merge `release.yml` into `ci.yml` so there is a single workflow file.

**Current flow:**
- `ci.yml`: runs on push to main / PR to main → lint + test + build (no deploy)
- `release.yml`: runs on tag push → build images + push to GHCR + deploy (no tests)

**New flow:**
- `ci.yml`: runs on push to main / PR to main → lint + test + build
- `ci.yml`: runs on tag push → same lint + test + build, THEN build images + push to GHCR + deploy

This improves on the current setup: production deploys now require all tests to pass first.

---

### Trigger change

```yaml
on:
  push:
    branches: [main]
    tags: ['v*.*.*']
  pull_request:
    branches: [main]
```

---

### New jobs to add (from release.yml, unchanged except for `needs` and `if`)

Both image jobs need:
- `if: startsWith(github.ref, 'refs/tags/')`
- `needs: [test-go, test-frontend, lint-go, lint-frontend]`
- `permissions: contents: read / packages: write`

`build-and-push-backend`:
```yaml
build-and-push-backend:
  name: Build and push backend image
  runs-on: ubuntu-latest
  if: startsWith(github.ref, 'refs/tags/')
  needs: [test-go, test-frontend, lint-go, lint-frontend]
  permissions:
    contents: read
    packages: write
  steps:
    - uses: actions/checkout@v5
    - name: Log in to GHCR
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ghcr.io/${{ github.repository_owner }}/homelab-backend
        tags: |
          type=semver,pattern={{version}}
          type=sha
          type=raw,value=latest
    - name: Build and push
      uses: docker/build-push-action@v6
      with:
        context: .
        file: docker/backend.Dockerfile
        target: prod
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
```

`build-and-push-frontend`: same pattern but for frontend image.

`deploy`:
```yaml
deploy:
  name: Deploy to homelab
  runs-on: self-hosted
  if: startsWith(github.ref, 'refs/tags/')
  needs: [build-and-push-backend, build-and-push-frontend]
  environment: homelab
  steps:
    - name: Deploy
      env:
        GHCR_TOKEN: ${{ secrets.GHCR_TOKEN }}
        GHCR_USER: ${{ github.repository_owner }}
      run: |
        echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin
        cd ~/homelab
        docker compose pull
        docker compose up -d
        docker image prune -f
        docker logout ghcr.io
        echo "Deploy done: $(date)"
```

---

### Remove

Delete `.github/workflows/release.yml` entirely.
