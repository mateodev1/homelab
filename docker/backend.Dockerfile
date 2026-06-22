# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Stage: base — download dependencies
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS base

WORKDIR /app

COPY go.work go.work.sum ./
COPY backend/go.mod backend/go.sum ./backend/
COPY cli/go.mod ./cli/
COPY shared/go.mod ./shared/
# cli and shared have no external deps — create empty go.sum so go.work resolves
RUN touch cli/go.sum shared/go.sum

RUN go mod download

# ---------------------------------------------------------------------------
# Stage: dev — hot reload with air
# ---------------------------------------------------------------------------
FROM base AS dev

COPY . .

RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", "backend/.air.toml"]

# ---------------------------------------------------------------------------
# Stage: build — compile static binary
# ---------------------------------------------------------------------------
FROM base AS build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./backend/cmd/api

# ---------------------------------------------------------------------------
# Stage: prod — minimal runtime image
# ---------------------------------------------------------------------------
FROM alpine:3.20 AS prod

RUN adduser -D -u 1000 appuser

COPY --from=build /app/bin/api /usr/local/bin/api

USER appuser

EXPOSE 8080

CMD ["/usr/local/bin/api"]
