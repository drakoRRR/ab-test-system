# AB Test Platform — CLAUDE.md

## Project Overview

A/B testing and feature flag management platform.

**Goal:** Build a self-hosted platform that enables:
- Feature flag management with targeting and rollout
- A/B/n experiment lifecycle management
- Deterministic user assignment (stable, hash-based)
- Exposure and conversion event collection
- Statistical analysis of experiment results (z-test, CI, uplift)
- Go SDK for client integration
- Demo app + k6 traffic generator for validation

## Monorepo Layout

```
ab-test-system/
├── backend/          # Main Go service (control plane + data plane)
│   ├── api/          # OpenAPI specs (modular YAML + Redocly)
│   ├── build/        # Dockerfiles
│   ├── cmd/server/   # Entry point
│   ├── config/       # YAML configs per environment
│   ├── internal/     # Application code (not importable externally)
│   │   ├── domain/   # Pure domain models (no deps)
│   │   ├── handler/  # HTTP handlers (oapi-codegen generated interfaces)
│   │   ├── infra/    # External integrations (GCP, DB)
│   │   └── services/ # Business logic
│   └── pkg/          # Shared utilities (importable)
├── frontend/         # Next.js admin UI
│   ├── app/          # App Router pages
│   ├── components/   # Shared UI components (shadcn/ui)
│   └── lib/          # API client, utils
├── sdk/              # Go SDK library (separate module)
│   └── go.mod
├── demo/             # Demo app that uses the SDK
│   └── go.mod
└── k6/               # Traffic generator scripts
    └── scenarios/
```

## Tech Stack

### Backend
| Component | Choice |
|---|---|
| Language | Go 1.23 |
| HTTP Router | gorilla/mux |
| API Codegen | oapi-codegen |
| Database | PostgreSQL (Docker local, Cloud SQL prod) |
| Migrations | golang-migrate |
| Message queue | Kafka 3.7 (KRaft, no Zookeeper) |
| Kafka Go client | github.com/segmentio/kafka-go |
| Cache / Pub/Sub / Rate limit | Redis 7.2 |
| Redis Go client | github.com/redis/go-redis/v9 |
| Auth | Firebase Auth (JWT verification) |
| Logging | zerolog |
| Config | YAML (gopkg.in/yaml.v3) |
| API Spec | OpenAPI 3.0.3 + Redocly CLI |
| Linting | golangci-lint |
| Testing | testify + testcontainers-go |
| Traffic gen | k6 |
| Containers | Docker + Docker Compose |

### Frontend
| Component | Choice |
|---|---|
| Framework | Next.js 14+ (App Router) + TypeScript |
| UI components | shadcn/ui + Radix UI |
| Styling | Tailwind CSS |
| Charts | Recharts |
| Auth | Firebase Auth (client SDK, Google login) |
| API client | fetch + generated types from OpenAPI |
| State | React Query (server state), Zustand (client state) |

## Architecture Decisions

### Deterministic User Assignment
`bucket = MurmurHash3(userID + ":" + experimentKey) % 10000`
- Same user always lands in the same bucket → stable assignment across requests
- No DB lookup needed for evaluation → low latency

### Event Model
Two event types tracked:
- `exposure` — user was assigned to a variant (logged by SDK on evaluation)
- `conversion` — user completed the goal action (logged by SDK explicitly)

### Statistical Analysis
- Metric: conversion rate per variant (conversions / exposures)
- Test: two-proportion z-test
- Output: p-value, 95% confidence interval, uplift %, statistical significance flag

### SDK Design
- Initialized with API key + service URL
- Polls the server for flag/experiment configs (TTL cache, default 30s)
- Evaluation is local (no network call per user)
- Events batched and flushed periodically

## Development Workflow

```bash
# Start dependencies
docker compose up -d

# Run migrations
make migrate-up

# Bundle OpenAPI specs
make bundle-api

# Generate Go code from spec
make generate-api

# Run server
make run

# Lint
make lint

# Test
make test
```

## Phase Roadmap

### Phase 1 — Foundation
- [ ] Docker Compose (PostgreSQL + Kafka KRaft + Kafka UI)
- [ ] DB migrations (users, organizations, projects)
- [ ] Auth implementation (Firebase → upsert user → profile)
- [ ] Projects CRUD
- [ ] SDK Keys management

### Phase 2 — Core Platform
- [ ] Feature Flags CRUD + evaluation engine
- [ ] Deterministic hashing (MurmurHash3)
- [ ] Experiments CRUD + lifecycle (draft → running → paused → completed)
- [ ] SDK Evaluate endpoint (`POST /sdk/evaluate`)

### Phase 3 — Events & Analytics
- [ ] Events ingestion: HTTP handler → Kafka producer (`POST /sdk/events` → 202)
- [ ] Events consumer: Kafka → PostgreSQL bulk insert + deduplication
- [ ] Metrics aggregation (SQL)
- [ ] Statistical analysis module (z-test, CI, p-value, uplift)
- [ ] Analytics API

### Phase 4 — Frontend (Admin UI)
- [ ] Next.js project setup (shadcn/ui, Tailwind, Firebase Auth)
- [ ] Auth flow (Google login → JWT → API calls)
- [ ] Projects list + create
- [ ] Feature Flags management (list, create, toggle kill switch)
- [ ] Experiments management (list, create, lifecycle controls)
- [ ] Analytics dashboard (conversion rates, uplift, p-value, significance badge)
- [ ] Timeseries chart (Recharts) per experiment
- [ ] SDK Keys management page

### Phase 5 — SDK & Demo
- [ ] Go SDK library (`sdk/` module)
- [ ] Demo app (`demo/` module)
- [ ] k6 traffic generator scripts

### Phase 6 — Polish
- [ ] Unit tests (hashing, statistics, middleware)
- [ ] Integration tests (DB)
- [ ] Full Docker Compose stack (backend + frontend + postgres)

## Key Patterns

- **Repository pattern**: `Storage` interface in service layer, PostgreSQL implementation in `infra/`
- **OpenAPI-first**: define spec → codegen → implement handler interface
- **Domain isolation**: domain models have zero external dependencies
- **Error handling**: wrap errors with context using `fmt.Errorf("op: %w", err)`, never swallow
- **Config**: load once at startup, pass via struct (no global config)

## API Module Files

```
api/
├── base.yml          # Common schemas, security, servers
├── auth.yml          # Auth endpoints
├── projects.yml      # Projects CRUD
├── flags.yml         # Feature flags
├── experiments.yml   # Experiments lifecycle
├── sdk.yml           # SDK evaluate + events
├── analytics.yml     # Statistics + timeseries
└── redocly.yaml      # Bundler config
```

## Database Schema (planned)

```
users               — id, email, name, photo_url, role, org_id, created_at, updated_at
organizations       — id, name, created_at
projects            — id, org_id, name, description, created_at, updated_at
sdk_keys            — id, project_id, key_hash, name, created_at, revoked_at
flags               — id, project_id, key, name, enabled, rules (jsonb), created_at, updated_at
experiments         — id, project_id, flag_id, name, status, traffic_percent, created_at, started_at, ended_at
variants            — id, experiment_id, key, name, weight
metrics             — id, experiment_id, event_name, is_primary
events              — id, project_id, user_id, experiment_id, variant_id, event_type, value, ts
```
