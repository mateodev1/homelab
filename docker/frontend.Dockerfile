# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Stage: base — install pnpm and fetch dependencies
# ---------------------------------------------------------------------------
FROM node:20-alpine AS base

RUN npm install -g pnpm

WORKDIR /app

COPY frontend/package.json frontend/pnpm-lock.yaml ./

# ---------------------------------------------------------------------------
# Stage: dev — development server with hot reload
# ---------------------------------------------------------------------------
FROM base AS dev

RUN pnpm install

COPY frontend/ .

CMD ["pnpm", "dev", "--host"]

# ---------------------------------------------------------------------------
# Stage: build — compile static assets
# ---------------------------------------------------------------------------
FROM base AS build

ARG VITE_AUTH0_DOMAIN
ARG VITE_AUTH0_CLIENT_ID

ENV VITE_AUTH0_DOMAIN=$VITE_AUTH0_DOMAIN
ENV VITE_AUTH0_CLIENT_ID=$VITE_AUTH0_CLIENT_ID

RUN pnpm install --frozen-lockfile

COPY frontend/ .

RUN pnpm build

# ---------------------------------------------------------------------------
# Stage: prod — minimal nginx image to serve static files
# ---------------------------------------------------------------------------
FROM nginx:alpine AS prod

COPY --from=build /app/dist /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
