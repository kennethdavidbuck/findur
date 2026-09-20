# Findur

Findur is an early-stage application. This repository currently contains a deployment walking skeleton: a PostgreSQL-backed Go API and a React status shell deployed behind one public Render Static Site origin. It deliberately does not include product, OAuth, or SnapTrade behavior yet.

## Repository layout

- `backend/` — Go API, migration command, and SQL migrations.
- `frontend/` — React, Vite, and TypeScript status shell.
- `render.yaml` — free Render web service, static site, and 30-day PostgreSQL database.
- `scripts/smoke-deployment.sh` — single-origin deployed smoke test.

Runtime versions are recorded in `.tool-versions`. Go and npm dependencies are locked by `backend/go.sum` and `frontend/package-lock.json`.

## Run locally

Start a local PostgreSQL instance and provide its connection string. No `.env` file is required or tracked.

```sh
export DATABASE_URL='postgresql://postgres:postgres@localhost:5432/findur?sslmode=disable'
cd backend
go run ./cmd/migrate
go run ./cmd/findur
```

In another terminal:

```sh
cd frontend
npm ci
npm run dev
```

Vite serves the shell on `http://localhost:5173` and proxies `/api/*` to the Go process on port `10000`.

## Verify

```sh
cd backend
go test ./...
go vet ./...
go build ./cmd/findur ./cmd/migrate
```

With `DATABASE_URL` set, repeat the migration to prove the no-change path:

```sh
cd backend
go run ./cmd/migrate
go run ./cmd/migrate
```

Verify the browser shell from the repository root:

```sh
npm --prefix frontend ci
npm --prefix frontend test -- --run
npm --prefix frontend run typecheck
npm --prefix frontend run lint
npm --prefix frontend run build
```

## Deploy to Render

Create a Blueprint from the root `render.yaml`. It defines `findur-api-kdb`, `findur-web-kdb`, and a free `findur-db` in Virginia (the static site is served globally). The web service receives the database's private connection string, runs migrations before starting, exposes `/api/healthz` for process liveness, and exposes PostgreSQL-backed `/api/readyz` for readiness. Both Git-backed services use `autoDeployTrigger: checksPass` so the authoritative GitHub checks gate automatic deployment.

Render free PostgreSQL databases expire after 30 days and are not production resources. Create the Blueprint at the start of the intended evaluation window.

After deployment, verify the static-site origin and its API rewrite without emitting request or response headers:

```sh
./scripts/smoke-deployment.sh https://YOUR-STATIC-SITE.onrender.com
```

The static rewrite intentionally targets the reserved `findur-api-kdb.onrender.com` hostname. If Render rejects that globally unique name or does not preserve the required rewrite behavior, capture the evidence and use the approved fallback: build the frontend and serve it from the Go web service. Do not introduce credentialed cross-origin CORS.

This skeleton proves only health, readiness, migrations, graceful shutdown, CI, and the initial one-origin deployment path. It does not complete the larger proxy/cookie or SnapTrade gates.
