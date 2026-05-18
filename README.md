# ShopVault

A simple e-commerce vulnerable web application built with Go and React.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Quick Start

```bash
git clone git@github.com:v0lka/ShopVault.git
cd ShopVault
docker compose up --build
```

The backend image is built with a multi-stage Dockerfile: a `golang:1.23-alpine` builder compiles the CGO-enabled binary, and the runtime stage is a slim `alpine:3.20` image containing only `sqlite-libs`, `imagemagick`, and `ca-certificates`. The compiled binary runs as a non-root `app` user.

### Development mode (hot reload)

The production compose file no longer bind-mounts source code. For live reload during development, use a compose override:

```yaml
# docker-compose.dev.yml
services:
  backend:
    image: golang:1.23-alpine
    volumes:
      - ./backend:/app
    working_dir: /app
    command: sh -c "apk add --no-cache gcc musl-dev sqlite-dev imagemagick && go run ./cmd/server"
```

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

## Access

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080

## Default Accounts

| Role  | Email                | Password    |
| ----- | -------------------- | ----------- |
| Admin | admin@shopvault.com  | admin123    |
| User  | customer@example.com | customer123 |

## Tech Stack

- **Backend**: Go 1.23, Gin framework, SQLite (CGO via `mattn/go-sqlite3`), ImageMagick (`convert`) for thumbnails
- **Frontend**: React 18, TypeScript, Vite, Bootstrap 5
- **Deployment**: Docker Compose, multi-stage Alpine images

## API Endpoints

See source code in `backend/cmd/server/main.go` for the full route table.
