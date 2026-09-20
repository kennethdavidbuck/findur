# Findur

Findur is an early-stage application. This repository contains a PostgreSQL-backed Go API and React frontend whose production artifacts independently carry the exact Git revision used to build them. It deliberately does not include product, OAuth, or live SnapTrade behavior yet.

## Repository layout

- `backend/` — Go API, migration command, and SQL migrations.
- `frontend/` — React, Vite, and TypeScript status shell.
- `render.yaml` — free Render web service, static site, and 30-day PostgreSQL database.
- `scripts/smoke-deployment.sh` — single-origin deployed smoke test.
- `compose.yaml` — production-shaped PostgreSQL, WireMock, API, frontend proxy, and disposable integration runner.
- `test/fixtures/wiremock/` — synthetic provider/proxy fixtures; never add live credentials or financial data.

Runtime versions are recorded in `.tool-versions`. Go and npm dependencies are locked by `backend/go.sum` and `frontend/package-lock.json`.

## Run locally

The shortest production-like workflow uses containers. The default stack builds both artifacts with the same full revision, migrates PostgreSQL before API admission, and binds the frontend only to loopback:

```sh
./scripts/compose-up.sh
```

This keeps the stack attached and streams service logs. Press `Ctrl+C` to stop it. If the stack was started in the background, reconnect to its logs with `./scripts/compose-logs.sh`.

Compose uses a deterministic synthetic build identity by default for local development and tags the resulting application images with it. CI overrides that value with the real full Git SHA; neither path creates a mutable application tag.

Open `http://localhost:8080`. The unlinked deployment diagnostic is at `http://localhost:8080/__status`.

For rebuild/restart development with Compose watch:

```sh
./scripts/compose-watch.sh
```

To run the complete synthetic integration suite over service DNS, including an actual headless Chromium execution of `/__status`:

```sh
./scripts/compose-test.sh
./scripts/compose-down.sh
```

Pass `--volumes` to `./scripts/compose-down.sh` when you intentionally want to reset local PostgreSQL data.

No Docker socket, Docker-in-Docker, Testcontainers, live provider endpoint, or real account data is used.

For the native workflow, start PostgreSQL and provide its connection string. No `.env` file is required or tracked.

```sh
export DATABASE_URL='postgresql://postgres:postgres@localhost:5432/findur?sslmode=disable'
cd backend
go run ./cmd/findur
```

The API applies pending migrations once before opening readiness; the migration library's tested no-change result handles an already-current database.

In another terminal:

```sh
cd frontend
npm ci
npm run dev
```

Vite serves the shell on `http://localhost:5173` and proxies `/api/*` to `VITE_API_PROXY` (default `http://localhost:10000`).

## Verify

```sh
cd backend
go test ./...
go vet ./...
go build ./cmd/findur ./cmd/migrate
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

Create a Blueprint from the root `render.yaml`. It defines `findur-api-kdb`, `findur-web-kdb`, and a free `findur-db` in Virginia. Independent Render auto-deploy is disabled: after every main-branch check passes, GitHub Actions publishes the backend's `linux/amd64` image to the public repository `docker.io/kdbuck/findur` under the full Git SHA, deploys that image's immutable digest through the Render API, invokes the protected frontend deploy hook for the same commit, and waits for public evidence that both surfaces report the same revision. Configure `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `RENDER_API_TOKEN`, `RENDER_BACKEND_SERVICE_ID`, `RENDER_FRONTEND_DEPLOY_HOOK_URL`, and `PUBLIC_ORIGIN` as GitHub Actions secrets. Normal releases publish no mutable application tag.

One-time Render provisioning requires an explicit bootstrap operation. Before the first Blueprint sync, build and verify a backend image from a real full Git SHA, push it to `docker.io/kdbuck/findur:<FULL_SHA>`, then create the `bootstrap-once` tag for that exact digest once. Do not move or reuse that tag. Render pulls the public image without registry credentials; the Docker Hub secrets are used only by GitHub Actions to authenticate publication. After provisioning, every normal release pushes the already integration-tested Compose image under its full-SHA tag and tells Render to deploy the resulting digest; normal releases never rebuild that backend artifact, edit `render.yaml`, or deploy `bootstrap-once`.

After the first full-SHA image passes the Compose integration gate, provision the bootstrap reference exactly once:

```sh
docker tag "findur-backend:$FULL_SHA" "docker.io/kdbuck/findur:$FULL_SHA"
docker push "docker.io/kdbuck/findur:$FULL_SHA"
docker tag "findur-backend:$FULL_SHA" "docker.io/kdbuck/findur:bootstrap-once"
docker push "docker.io/kdbuck/findur:bootstrap-once"
```

Render free PostgreSQL databases expire after 30 days and are not production resources. Create the Blueprint at the start of the intended evaluation window.

After deployment, verify the static-site origin and its API rewrite without emitting request or response headers:

```sh
./scripts/smoke-deployment.sh https://YOUR-STATIC-SITE.onrender.com "$(git rev-parse HEAD)"
```

The static rewrite intentionally targets the reserved `findur-api-kdb.onrender.com` hostname. If Render rejects that globally unique name or does not preserve the required rewrite behavior, capture the evidence and use the approved fallback: build the frontend and serve it from the Go web service. Do not introduce credentialed cross-origin CORS.

The smoke is bounded and fails for unavailable, stale, malformed, or mismatched frontend/API identities. Health and status checks do not make provider calls. The tracked WireMock fixtures contain only categorical synthetic values.
