# SplitLab Backend

The control plane and data plane for the SplitLab A/B testing and feature flag platform.

## Architecture Overview

```
┌──────────────┐     Firebase JWT      ┌──────────────────────────────────┐
│  Admin UI    │ ──────────────────►   │  HTTP API  (:8080/api/v1)        │
│  (Next.js)   │                       │  gorilla/mux + oapi-codegen      │
└──────────────┘                       │                                  │
                                       │  Control plane   SDK plane       │
┌──────────────┐     X-API-Key         │  /projects       /sdk/config     │
│  Go SDK      │ ──────────────────►   │  /flags          /sdk/events     │
│  (clients)   │                       │  /experiments    /sdk/evaluate   │
└──────────────┘                       │  /analytics                      │
                                       └───────────┬──────────────────────┘
                                                   │
                          ┌────────────────────────┼────────────────────┐
                          ▼                        ▼                    ▼
                    PostgreSQL 16            Kafka 3.7             Redis 7.2
                    (primary store)      (event pipeline)       (config cache)
                          ▲
                    ┌─────┴──────┐
                    │  Consumer  │
                    │  (worker)  │  Kafka → PostgreSQL bulk insert
                    └────────────┘
```

Two long-running processes:

| Process | Entry point | Responsibility |
|---|---|---|
| `server` | `cmd/server/main.go` | HTTP API (control plane + SDK plane) |
| `consumer` | `cmd/consumer/main.go` | Kafka consumer → events table bulk insert |

## Getting Started

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- A Firebase project (for admin auth)
- `redocly` CLI — `npm i -g @redocly/cli` (only needed to regenerate code from spec)
- `oapi-codegen` — `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest` (same)

### One-command dev start

```bash
cd backend
make dev
```

This:
1. Starts PostgreSQL, Kafka (KRaft), Redis, and Kafka UI via Docker Compose
2. Waits for health checks to pass
3. Runs all pending DB migrations
4. Starts `server` and `consumer` in parallel (Ctrl+C kills both)

The API is available at `http://localhost:8080/api/v1`.

### Step-by-step

```bash
# 1. Start infrastructure
make compose-up

# 2. Run migrations
make migrate-up

# 3. Start the HTTP server
make run

# 4. Start the event consumer (separate terminal)
make run-consumer
```

### Configuration

The server reads a YAML config file. The path is passed via `-config` flag (default: `./config/example.yaml`).

```yaml
# config/example.yaml
http_server:
  port: 8080

gcp_auth:
  project_id: "your-firebase-project-id"

database:
  dsn: "postgres://abtest:abtest@localhost:5432/abtest?sslmode=disable"
  max_conns: 10
  min_conns: 2

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

kafka:
  brokers:
    - "localhost:9092"
  producers:
    events:
      topic: "ab-test-events"
  consumers:
    events:
      topic: "ab-test-events"
      group_id: "ab-test-events-consumer"
      batch_size: 100
      flush_timeout: 2s
```

## API

The API is defined in OpenAPI 3.0 YAML specs under `api/`. All routes are prefixed with `/api/v1`.

### Authentication

| Route group | Auth mechanism |
|---|---|
| Control plane (`/projects`, `/flags`, `/experiments`, `/analytics`) | Firebase JWT — `Authorization: Bearer <token>` |
| SDK plane (`/sdk/*`) | API key — `X-API-Key: <sdk_key>` |

### Control Plane Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/users/me` | Get or create the authenticated user's profile |
| `POST` | `/organizations` | Create an organisation |
| `GET/POST` | `/projects` | List / create projects |
| `GET/PATCH/DELETE` | `/projects/{id}` | Read / update / delete a project |
| `GET/POST` | `/projects/{id}/flags` | List / create feature flags |
| `GET/PATCH/DELETE` | `/projects/{id}/flags/{key}` | Read / update / delete a flag |
| `GET/POST` | `/projects/{id}/experiments` | List / create experiments |
| `GET/PATCH` | `/projects/{id}/experiments/{expId}` | Read / update an experiment |
| `POST` | `/projects/{id}/experiments/{expId}/start` | Transition `draft → running` |
| `POST` | `/projects/{id}/experiments/{expId}/pause` | Transition `running → paused` |
| `POST` | `/projects/{id}/experiments/{expId}/complete` | Transition `→ completed` |
| `GET` | `/projects/{id}/experiments/{expId}/analytics` | Statistical results (uplift, p-value, CI) |
| `GET/POST` | `/projects/{id}/sdk-keys` | List / create SDK keys |
| `DELETE` | `/projects/{id}/sdk-keys/{keyId}` | Revoke an SDK key |

### SDK Plane Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/sdk/config` | Pull flag + experiment config snapshot for the SDK |
| `POST` | `/sdk/events` | Ingest exposure / conversion events (→ Kafka) |

## Project Layout

```
backend/
├── api/                    # OpenAPI specs (modular YAML)
│   ├── base.yml            # common schemas, security definitions
│   ├── auth.yml, projects.yml, flags.yml, experiments.yml ...
│   ├── sdk.yml             # SDK-specific endpoints
│   ├── analytics.yml
│   └── _build/             # generated bundles (gitignored)
├── cmd/
│   ├── server/main.go      # HTTP server entry point
│   ├── consumer/main.go    # Kafka consumer entry point
│   └── migrate/main.go     # migration runner
├── config/
│   └── example.yaml
├── internal/
│   ├── domain/             # pure domain models, errors — no external deps
│   ├── services/           # business logic, repository interfaces
│   ├── handler/http/       # HTTP handlers (oapi-codegen strict server)
│   └── infra/              # PostgreSQL, Kafka, Redis, Firebase implementations
├── migrations/             # SQL migration files
└── pkg/                    # shared utilities
```

## Database Schema

```
organizations   id, name, created_at
users           id, email, name, photo_url, role, org_id, created_at, updated_at
projects        id, org_id, name, description, created_at, updated_at
sdk_keys        id, project_id, key_hash, name, created_at, revoked_at
flags           id, project_id, key, name, enabled, rules (jsonb), created_at, updated_at
experiments     id, project_id, key, name, status, traffic_percent, started_at, ended_at
variants        id, experiment_id, key, name, weight
events          id, project_id, user_id, experiment_id, variant_id, type, name, value, ts
```

## Development

### Run tests

```bash
make test
```

Unit tests use testify mocks (generated by mockery). Integration tests use testcontainers to spin up real Postgres.

### Lint

```bash
make lint
```

Runs `gofumpt` + `goimports` + `golangci-lint`.

### Regenerate API code

After editing any `api/*.yml` file:

```bash
make generate-api
```

This bundles all specs and regenerates `internal/handler/http/api/codegen/api.gen.go`.
The project will not compile until every new `operationId` in the spec has a corresponding handler method.

### Regenerate mocks

```bash
make generate-mocks
```

Uses mockery with the config in `mockery.yaml`. Never edit `*_gen.go` files by hand.

## Statistical Analysis

The analytics endpoint computes experiment results using a **two-proportion z-test**:

- **Conversion rate** per variant = conversions / exposures
- **Uplift** = (treatment CR − control CR) / control CR × 100%
- **p-value** from two-tailed z-test on proportions
- **95% confidence interval** on the difference
- **Significance flag** = p-value < 0.05

Results are available at `GET /projects/{id}/experiments/{expId}/analytics` once the experiment has accumulated events.
