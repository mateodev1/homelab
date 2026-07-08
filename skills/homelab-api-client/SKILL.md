---
name: homelab-api-client
description: "Trigger: API client, connect API, endpoints, api key, homelab api. Connect to the homeLab Todo API from any external client using two env vars."
license: MIT
metadata:
  author: mateo
  version: "1.0"
---

## Activation Contract

Use this skill when writing code that connects to the homeLab Todo API as an external HTTP client — any language, any environment, any deployment target.

## Hard Rules

- Configure the client with exactly two environment variables: `API_BASE_URL` (e.g. `https://homelab.example.com`) and `API_KEY`.
- Send `Authorization: Bearer <API_KEY>` on every request. **Enforced by the server when `ENV=production`.** In development (`ENV=development` or unset) the server accepts unauthenticated requests, but sending the header is still recommended so the client works unchanged against production. `/api/health` never requires the key.
- All request/response bodies are JSON. Set `Content-Type: application/json` on requests with a body.
- `priority` is an integer `0-3`. `status` is one of `todo | in_progress | done | cancelled`.
- Full endpoint shapes live in `references/endpoints.md` — read it before implementing calls.
- Do not hardcode the base URL or key; always read them from `API_BASE_URL` / `API_KEY`.

## Decision Gates

| Situation | Action |
| --- | --- |
| Building a new client (any language) | Read `references/endpoints.md`, implement all 5 endpoints, use env vars for config |
| Talking to a production deployment | Header is required — requests without it get `401 {"error": "unauthorized"}` |
| Unsure of a field name or enum value | Check `references/endpoints.md`, do not guess |

## Execution Steps

1. Read `API_BASE_URL` and `API_KEY` from the environment at startup; fail fast if either is missing.
2. Read `references/endpoints.md` for exact request/response shapes.
3. Implement calls with `Authorization: Bearer <API_KEY>` on every request.
4. Handle standard HTTP status codes: `200`, `201`, `204`, `400`, `401`, `404`, `500` (error bodies are `{"error": "message"}`).

## Output Contract

Return which endpoints were implemented and confirm the two env vars are used exclusively for configuration.

## References

- `references/endpoints.md` — full endpoint list, request/response JSON shapes, status/priority enums.
