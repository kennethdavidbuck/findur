# Findur

> Meet the portfolio before the profile photo.

Findur flips the familiar dating-app script. Instead of leading with a headshot and hoping financial compatibility turns up later, it starts with the investing signals a person has chosen to share. Someone might look for a similar investing style, a complementary one, or the kind of diversification that sparks a good first conversation. Candidate experiences put compatibility before appearance, and photos stay obscured until the interest is mutual.

The idea is playful; the data boundary is serious. A connected portfolio can reveal something interesting about investing style, but it is not a complete financial identity. Findur does not equate holdings with net worth, certify financial responsibility, rank human worth, provide investment advice, or expose one person's raw financial data to another. Connection, private matching inputs, visible disclosure, and withdrawal of consent remain separate choices.

This repository develops a responsive demonstration built around one protected owner, SnapTrade's test OAuth environment, and synthetic candidate data. It is a product and engineering proof—not a public dating launch.

## From connection to chemistry

The planned end-to-end experience begins with a secure, consent-led portfolio connection. The owner chooses what Findur may use, sees how portfolio data becomes private compatibility signals, and decides what belongs on their profile. From there, synthetic candidates bring the concept to life through portfolio-first discovery, swiping, and a mutual-match photo reveal. The goal is still chemistry—Findur simply changes the opening question.

Delivery is organized into epics and stories rather than described as a fixed snapshot in this README. The [epics and stories](_bmad-output/planning-artifacts/epics.md) define the roadmap and acceptance criteria. Story-level implementation specs and review evidence live in [`_bmad-output/implementation-artifacts/`](_bmad-output/implementation-artifacts/), while the living [sprint status](_bmad-output/implementation-artifacts/sprint-status.yaml) shows what is complete, active, or still planned.

## How the idea becomes a product

Findur is being planned and built with the BMad Method. The repository keeps the path from “what if?” to working software visible, so the product thinking can be explored alongside the code instead of disappearing into a separate planning system.

| Stage | Artifact | What it establishes |
| --- | --- | --- |
| Ask the big question | [Product brief](_bmad-output/planning-artifacts/briefs/brief-findur-2026-09-19/brief.md) | The problem, audience, product promise, trust boundary, and initial vision. |
| Test it against reality | [SnapTrade integration research](_bmad-output/planning-artifacts/research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/research.md) | Provider capabilities, OAuth and data constraints, operational risks, and the safe demonstration boundary. |
| Turn the idea into promises | [Product requirements](_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/prd.md) | User journeys, functional requirements, non-goals, success measures, and privacy and safety requirements. |
| Design the moments that matter | [Experience spine](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md) and [design spine](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/DESIGN.md) | Information architecture, interaction states, accessibility, responsive behaviour, visual language, and key flows. |
| Give it strong bones | [Architecture spine](_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/ARCHITECTURE-SPINE.md) and [data model](_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/DATA-MODEL.md) | System boundaries, security decisions, consistency rules, deployment shape, and data ownership. |
| Ship it in meaningful slices | [Epics and stories](_bmad-output/planning-artifacts/epics.md) | The implementation path from secure connection through portfolio-led discovery and matching. |

Together, these artifacts preserve the chain from product intent to acceptance criteria, implementation, review, and delivery status.

### What is captured inside

- **Research that can be audited:** the feasibility report is backed by a structured [claims ledger](_bmad-output/planning-artifacts/research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/claims.json), focused source digests, and a [citation-verification record](_bmad-output/planning-artifacts/research/technical-snaptrade-commercial-integration-feasibi-2026-09-19/digests/citation-verification.md). Recommendations distinguish documented provider behaviour from product inference and unresolved policy questions.
- **Requirements with a visible decision trail:** the PRD carries stable requirement IDs, assumptions, non-goals, and source notes. Representative reconciliation records show how the [brief](_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/reconcile-brief.md), [SnapTrade research](_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/reconcile-snaptrade-research.md), and [UX experience](_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/reconcile-ux-experience.md) were brought into agreement; the [product-fidelity recheck](_bmad-output/planning-artifacts/prds/prd-findur-2026-09-19/review-product-fidelity-recheck.md) records how the earlier high findings were resolved.
- **UX beyond prose:** [Constellation](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/.working/direction-constellation.html) is the selected composition reference, with three alternative visual directions and five key-screen explorations retained in the [working HTML gallery](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/.working/); the final written spines govern whenever an exploration differs. Those spines define components and tokens, responsive phone/desktop behaviour, English/French and light/dark parity, interaction states, recovery paths, and a [WCAG 2.2 AA accessibility floor](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/EXPERIENCE.md#accessibility-floor). A historical [pre-remediation accessibility and trust review](_bmad-output/planning-artifacts/ux-designs/ux-findur-2026-09-19/review-accessibility-trust.md) preserves the gaps that drove the final revisions; its verdict does not describe the status of the finalized spines.
- **Architecture that records trade-offs:** the architecture spine contains adopted decisions for identity, encryption, provider isolation, refresh, data minimization, consistency, accessibility, deployment, and observability. Its reviews test technology currency and cross-boundary failure modes; the [adversarial seam review](_bmad-output/planning-artifacts/architecture/architecture-findur-2026-09-19/reviews/review-adversarial-seams.md#resolution-check) keeps both the original concerns and their resolution visible.
- **Delivery with memory:** epics map requirements into user-valued slices, implementation specs retain acceptance decisions and review triage, [deferred work](_bmad-output/implementation-artifacts/deferred-work.md) prevents consciously postponed concerns from disappearing, and [sprint status](_bmad-output/implementation-artifacts/sprint-status.yaml) provides the live view without freezing story progress into this README.

### A queryable project record

The documents for the whole project live here with the code as versioned Markdown, YAML, HTML, and generated evidence. They can be read directly, compared through Git history, or queried with a coding agent. For example, an agent can trace a trust decision from the product brief into a PRD requirement, find the architecture rule that supports it, identify the stories that deliver it, and inspect the corresponding implementation spec and status—all from repository evidence.

That makes the planning useful beyond its original sessions: a new contributor or reviewer can ask their own agent how a feature was shaped, what trade-offs were considered, what remains planned, and where the governing decision lives.

### AI-assisted workflow

BMad workflows were run through Codex, primarily with [GPT-5.6 Sol](https://developers.openai.com/api/docs/models/gpt-5.6-sol) (`gpt-5.6-sol`) at medium and high reasoning effort. The model was used as a planning, implementation, and review collaborator; repository artifacts, tests, generated contracts, and human direction remained the durable source of truth. No OpenAI model is part of the Findur application runtime.

## Take Findur for a spin

### Prerequisites

For the container workflow:

- Docker with Docker Compose v2
- Git

Native development additionally requires Go and Node.js/npm. Exact versions are pinned in [`.tool-versions`](.tool-versions).

### Start the complete local stack

From the repository root:

```sh
./scripts/compose-up.sh
```

This builds the application and starts PostgreSQL, WireMock, the Go API, and the production frontend proxy. Open [http://127.0.0.1:8080](http://127.0.0.1:8080). The deployment diagnostic is available at [http://127.0.0.1:8080/__status](http://127.0.0.1:8080/__status).

The local stack uses a complete synthetic OAuth and masked-account flow by default. Choosing **Continue to SnapTrade** completes authorization against WireMock without contacting SnapTrade or requiring real credentials or financial data.

Press `Ctrl+C` to stop the attached stack, then remove its containers and network with:

```sh
./scripts/compose-down.sh
```

PostgreSQL data is preserved between runs. To intentionally reset it, use `./scripts/compose-down.sh --volumes`.

For rebuild-and-restart development:

```sh
./scripts/compose-watch.sh
```

To run the complete synthetic integration suite, including the headless-browser OAuth and session flow:

```sh
./scripts/compose-test.sh
./scripts/compose-down.sh
```

### Run the application natively

Start the repository's PostgreSQL service and provide its matching connection string; no `.env` file is required or tracked.

```sh
docker compose up --wait postgres
export DATABASE_URL='postgresql://findur:synthetic-local-only@localhost:5432/findur?sslmode=disable'
cd backend
go run ./cmd/findur
```

In another terminal:

```sh
cd frontend
npm ci
npm run dev
```

Vite serves the frontend at [http://localhost:5173](http://localhost:5173) and proxies `/api/*` to `VITE_API_PROXY`, which defaults to `http://localhost:10000`.

Native and production environments keep SnapTrade authorization closed unless the feature gate, issuer, client credentials, callback, provider API origin, public origin, and independent hashing/encryption keys are supplied explicitly. The canonical environment names and validation rules live in [`backend/internal/platform/config/config.go`](backend/internal/platform/config/config.go). HTTP callbacks are accepted only for loopback hosts; non-loopback callbacks require HTTPS.

## Verification

Verify the Go API and generated provider boundary:

```sh
cd backend
go generate ./...
golangci-lint run
go test ./...
go vet ./...
go build ./cmd/findur ./cmd/migrate
```

Verify the browser application from the repository root:

```sh
npm --prefix frontend ci
npm --prefix frontend test -- --run
npm --prefix frontend run typecheck
npm --prefix frontend run lint
npm --prefix frontend run build
```

The checked-in SnapTrade provider client is generated from a checksum-pinned official OpenAPI document. The local OAuth bearer overlay selects only the approved read operations, removes Commercial credential query parameters, and applies OAuth bearer security before generation.

## Repository guide

- `backend/` — Go API, database migrations, OpenAPI contracts, and the generated SnapTrade provider boundary.
- `frontend/` — React, Vite, and TypeScript public and protected experiences.
- `_bmad-output/planning-artifacts/` — product, research, UX, architecture, and epic-level decisions.
- `_bmad-output/implementation-artifacts/` — implementation specs, review history, deferred work, and sprint status.
- `scripts/` — local Compose, integration, and deployed smoke-test commands.
- `test/fixtures/wiremock/` — categorical synthetic provider fixtures; never add live credentials or financial data.
- `compose.yaml` — the production-shaped local and integration environment.
- `render.yaml` — the Render Blueprint for the hosted demonstration.

## Deployment model

The hosted demonstration uses a Render static site, Go web service, and PostgreSQL database. CI builds and integration-tests the backend image, publishes it under the full Git SHA, deploys its immutable digest, deploys the frontend for the same commit, and verifies that both public surfaces report that revision. Render auto-deploy and Blueprint auto-sync remain disabled after initial provisioning so the CI workflow is the sole release path.

Deployment requires repository secrets for Docker Hub publication, Render API access, the protected frontend deploy hook, and the public origin. Do not commit credentials, `.env` files, OAuth tokens, provider secrets, real financial data, or unredacted logs.

The first Blueprint sync needs the one-time `bootstrap-once` image referenced by `render.yaml`. Build and verify a backend image from a real full Git SHA, then publish both references to the same image:

```sh
docker tag "findur-backend:$FULL_SHA" "docker.io/kdbuck/findur:$FULL_SHA"
docker push "docker.io/kdbuck/findur:$FULL_SHA"
docker tag "findur-backend:$FULL_SHA" "docker.io/kdbuck/findur:bootstrap-once"
docker push "docker.io/kdbuck/findur:bootstrap-once"
```

Create `bootstrap-once` only for provisioning; do not move or reuse it. Normal releases deploy the tested full-SHA image by immutable digest.

After deployment, verify the public origin without emitting request or response headers:

```sh
./scripts/smoke-deployment.sh https://YOUR-STATIC-SITE.onrender.com "$(git rev-parse HEAD)"
```

[Render's free PostgreSQL service expires 30 days after creation](https://render.com/docs/free), has no backups, and is not a production data store. Provision it for the intended demonstration window. The hosted environment is a demonstration boundary, not authorization for real-user or production financial data.
