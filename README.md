# SplitLab — A/B Testing & Feature Flag Platform

A self-hosted platform for managing feature flags and running A/B experiments.
Built as a diploma project at [University Name].

**Core capabilities:**

- Feature flag management with percentage rollout and kill switch
- A/B/n experiment lifecycle (draft → running → paused → completed)
- Deterministic, stable user assignment via MurmurHash3
- Exposure and conversion event collection through a Kafka pipeline
- Statistical analysis: two-proportion z-test, uplift %, 95% CI, p-value
- Go SDK for client integration with local evaluation and async event buffering
- Admin UI for managing flags, experiments, and reviewing analytics
- Demo app + k6 load generator for end-to-end validation

![Analytics dashboard showing control 10.0% vs treatment 11.6%, +16.5% uplift, p=0.0036](demo/media/results.png)

---

## Architecture

```
┌──────────────────┐  Firebase JWT   ┌─────────────────────────────────────┐
│   Admin UI       │ ─────────────►  │  Backend  :8080/api/v1              │
│   Next.js 14     │                 │  gorilla/mux  +  oapi-codegen       │
└──────────────────┘                 │                                     │
                                     │  Control plane    SDK plane         │
┌──────────────────┐  X-API-Key      │  /projects        /sdk/config       │
│   Go SDK         │ ─────────────►  │  /flags           /sdk/events       │
│   (embedded in   │                 │  /experiments                       │
│    your app)     │                 │  /analytics                         │
└──────────────────┘                 └────────────┬────────────────────────┘
                                                  │
                         ┌────────────────────────┼──────────────────────┐
                         ▼                        ▼                      ▼
                   PostgreSQL 16            Kafka 3.7               Redis 7.2
                   (primary store)      (event pipeline)         (config cache)
                         ▲
                   ┌─────┴──────┐
                   │  Consumer  │   Kafka → PostgreSQL bulk insert
                   └────────────┘
```

### Assignment algorithm

```
bucket = MurmurHash3_32(userID + ":" + experimentKey) % 10_000
```

The same formula runs in both the backend and the Go SDK — a user always lands in the same variant regardless of which side evaluates the assignment.

### Event flow

```
SDK.GetVariant()  →  exposure event  ─┐
SDK.Track()       →  conversion event ─┴─►  POST /sdk/events  ►  Kafka  ►  Consumer  ►  PostgreSQL
```

Analytics are computed on-demand from the raw events table using SQL aggregation + z-test.

---

## Monorepo Layout

```
ab-test-system/
├── backend/          # Go service — control plane + SDK plane + event consumer
├── frontend/         # Next.js 14 admin UI
├── sdk/              # Go SDK library (separate module)
├── demo/             # Demo storefront app that uses the SDK
│   ├── backend/      # Go HTTP server
│   └── static/       # Vanilla HTML frontend (no build step)
├── k6/               # k6 traffic generator scripts
│   └── scenarios/
└── Makefile          # Top-level convenience targets (demo-setup, demo-run, demo-k6)
```

Each component has its own README:

| Component | README |
|---|---|
| Backend | [backend/README.md](backend/README.md) |
| Go SDK | [sdk/README.md](sdk/README.md) |
| Demo app | [demo/README.md](demo/README.md) |

---

## Tech Stack

| Area | Technology                                      |
|---|-------------------------------------------------|
| Backend language | Go 1.25                                         |
| HTTP routing | gorilla/mux                                     |
| API definition | OpenAPI 3.0 + oapi-codegen                      |
| Database | PostgreSQL 16                                   |
| Migrations | golang-migrate                                  |
| Message queue | Kafka 3.7 (KRaft)                               |
| Cache | Redis 7.2                                       |
| Auth | Firebase Auth (Google login)                    |
| Frontend | Next.js 14, TypeScript, shadcn/ui, Tailwind CSS |
| Charts | Recharts                                        |
| SDK hashing | MurmurHash3-32 (inlined, no unsafe)             |
| Load testing | k6                                              |
| Containers | Docker + Docker Compose                         |

---

## Quick Start

### 1. Start the backend

```bash
cd backend
make dev          # starts infra, migrates, runs server + consumer
```

### 2. Start the admin UI

```bash
cd frontend
pnpm install
pnpm dev          # http://localhost:3000
```

### 3. Run the demo end-to-end

```bash
# Seed platform resources and write demo/.env
export FIREBASE_TOKEN=<token from DevTools → Cookies>
make demo-setup

# Start the demo store
make demo-run     # http://localhost:8081

# Generate synthetic traffic with k6
make demo-k6
```

See [demo/README.md](demo/README.md) for the full walkthrough.

---

## Statistical Analysis

Experiment results are computed using a **two-proportion z-test** on the raw exposure and conversion counts:

| Metric | Formula |
|---|---|
| Conversion rate | conversions / exposures |
| Uplift | (CR_treatment − CR_control) / CR_control |
| z-statistic | (p̂₁ − p̂₂) / SE(p̂₁ − p̂₂) |
| p-value | two-tailed from standard normal |
| 95% CI | (p̂₁ − p̂₂) ± 1.96 × SE |
| Significant | p-value < 0.05 |

---

## License

[MIT](LICENSE) © 2025 Vlad Musaelyan
