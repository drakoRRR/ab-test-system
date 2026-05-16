# System Architecture

## Overview

AB Test Platform is a self-hosted feature flag and A/B experiment management system.
It is divided into two planes:

- **Control Plane** — admin UI + backend API for managing flags, experiments, and analytics
- **Data Plane** — SDK endpoints: evaluate (assignment), events ingestion, config polling

---

## High-Level Diagram

```mermaid
graph TB
  subgraph "Control Plane"
    UI["Next.js Admin UI<br/>(shadcn/ui + Tailwind)"]
    UI -->|"HTTPS + Firebase JWT"| API
  end

  subgraph "Backend (Go)"
    API["HTTP Server<br/>gorilla/mux + oapi-codegen"]
    API -->|"verify token"| FB["Firebase Auth"]
    API -->|"produce events"| KF[("Kafka<br/>ab.events.raw")]
    API -->|"SQL (CRUD, analytics)"| DB[("PostgreSQL")]
    API -->|"cache-aside GET/SET"| RD[("Redis")]
    API -->|"PUBLISH invalidation"| RD
    API -->|"rate limit Lua script"| RD
    KF -->|"consume batch"| CONS["Events Consumer<br/>(goroutine)"]
    CONS -->|"bulk INSERT"| DB
    API --> STATS["Stats Engine<br/>(pure Go z-test)"]
    RD -->|"SUBSCRIBE invalidation"| INV["Cache Invalidator<br/>(goroutine)"]
    INV -->|"DEL config key"| RD
  end

  subgraph "Data Plane (SDK)"
    DEMO["Demo App (Go)"]
    SDK["Go SDK"]
    DEMO --> SDK
    SDK -->|"GET /sdk/config (poll 30s)"| API
    SDK -->|"POST /sdk/events → 202"| API
  end

  subgraph "Load Testing"
    K6["k6 scripts"]
    K6 -->|"simulate N users"| DEMO
  end

  subgraph "Observability"
    KUI["Kafka UI :8090"]
    KUI --> KF
  end
```

---

## Request Flows

### 1. Flag / Experiment Evaluation (hot path)

```mermaid
sequenceDiagram
  participant App as Demo App
  participant SDK as Go SDK
  participant Cache as Local Config Cache
  participant API as Backend API
  participant DB as PostgreSQL

  App->>SDK: EvaluateExperiment(userID, "checkout-btn")
  SDK->>Cache: get config for "checkout-btn"
  alt cache miss or expired (>30s)
    SDK->>API: GET /sdk/config
    API->>DB: SELECT active experiments/flags
    DB-->>API: configs
    API-->>SDK: ConfigResponse
    SDK->>Cache: store configs
  end
  Cache-->>SDK: ExperimentConfig{variants, traffic%}
  SDK->>SDK: bucket = MurmurHash3(userID+":checkout-btn") % 10000
  SDK->>SDK: variant = assignVariant(bucket, variants, traffic%)
  SDK-->>App: VariantKey ("control" | "treatment")

  Note over SDK: async, non-blocking
  SDK-)API: POST /sdk/events [{type:exposure, ...}]
```

### 2. Conversion Tracking (via Kafka)

```mermaid
sequenceDiagram
  participant App as Demo App
  participant SDK as Go SDK
  participant Buffer as Event Buffer (channel)
  participant API as Backend API
  participant KF as Kafka ab.events.raw
  participant CONS as Events Consumer
  participant DB as PostgreSQL

  App->>SDK: Track("purchase", userID, value=1)
  SDK->>Buffer: enqueue event{id: uuid, type:conversion, ...}
  Note over Buffer: flush: 100 events OR 1s

  Buffer->>API: POST /sdk/events [batch]
  API->>API: validate batch
  API->>KF: produce messages (key=experiment_id)
  API-->>Buffer: 202 Accepted

  Note over KF,CONS: async, decoupled
  loop 500 msgs OR 500ms
    KF->>CONS: fetch batch
    CONS->>DB: INSERT ... ON CONFLICT (event_id) DO NOTHING
    DB-->>CONS: ok
    CONS->>KF: commit offset
  end
```

### 3. Analytics Computation (on-demand)

```mermaid
sequenceDiagram
  participant UI as Admin UI
  participant API as Backend API
  participant DB as PostgreSQL
  participant SE as Stats Engine (Go)

  UI->>API: GET /projects/{pid}/experiments/{eid}/analytics
  API->>DB: SELECT variant_id, event_type, COUNT(*) FROM events WHERE experiment_id=? GROUP BY 1,2
  DB-->>API: [{variant:"control", exposures:5000, conversions:502}, {variant:"treatment", ...}]
  API->>SE: Compute(variants)
  SE->>SE: z-test for two proportions
  SE->>SE: 95% confidence interval (Wilson score)
  SE->>SE: uplift = (p_treatment - p_control) / p_control
  SE-->>API: AnalyticsResult{pValue, ci, uplift, isSignificant}
  API-->>UI: AnalyticsResponse
```

---

## Database Schema

```mermaid
erDiagram
  organizations {
    uuid id PK
    string name
    timestamp created_at
  }

  users {
    uuid id PK
    uuid org_id FK
    string email
    string name
    string photo_url
    enum role "admin|member|viewer"
    timestamp created_at
    timestamp updated_at
  }

  projects {
    uuid id PK
    uuid org_id FK
    string name
    string description
    timestamp created_at
    timestamp updated_at
  }

  sdk_keys {
    uuid id PK
    uuid project_id FK
    string key_hash
    string name
    timestamp created_at
    timestamp revoked_at
  }

  flags {
    uuid id PK
    uuid project_id FK
    string key
    string name
    bool enabled
    jsonb rules "targeting rules"
    timestamp created_at
    timestamp updated_at
  }

  experiments {
    uuid id PK
    uuid project_id FK
    uuid flag_id FK
    string name
    enum status "draft|running|paused|completed"
    int traffic_percent
    string hypothesis
    timestamp created_at
    timestamp started_at
    timestamp ended_at
  }

  variants {
    uuid id PK
    uuid experiment_id FK
    string key "control|treatment|..."
    string name
    int weight "relative weight, sum=100"
  }

  metrics {
    uuid id PK
    uuid experiment_id FK
    string event_name "purchase|signup|..."
    bool is_primary
  }

  events {
    uuid id PK "generated by SDK — idempotency key"
    uuid project_id FK
    string user_id "external, not UUID"
    uuid experiment_id FK
    uuid variant_id FK
    enum event_type "exposure|conversion"
    string event_name
    float value
    timestamp ts
  }

  organizations ||--o{ users : "has"
  organizations ||--o{ projects : "owns"
  projects ||--o{ sdk_keys : "has"
  projects ||--o{ flags : "has"
  projects ||--o{ experiments : "has"
  flags ||--o{ experiments : "drives"
  experiments ||--o{ variants : "has"
  experiments ||--o{ metrics : "measures"
  experiments ||--o{ events : "collects"
  variants ||--o{ events : "assigned to"
```

---

## Deployment (Docker Compose)

```mermaid
graph LR
  subgraph "docker compose"
    FE["frontend:3000<br/>(Next.js)"]
    BE["backend:8080<br/>(Go)"]
    PG["postgres:5432<br/>(PostgreSQL 16)"]
    KF["kafka:9092<br/>(KRaft, no ZK)"]
    KUI["kafka-ui:8090"]
    RD["redis:6379<br/>(redis:7.2-alpine)"]
    MG["migrate<br/>(one-shot)"]
  end

  FE -->|HTTP| BE
  BE -->|SQL| PG
  BE -->|produce| KF
  BE -->|consume| KF
  BE -->|cache / pubsub / ratelimit| RD
  KUI --> KF
  MG -->|migrations| PG

  K6["k6 (host)"] -->|HTTP| BE
```

---

## Key Constraints

| Constraint | Value | Rationale |
|---|---|---|
| SDK config cache TTL | 30s | Balance between freshness and DB load |
| Event batch size | max 500 events | Payload size limit |
| Event flush interval | 1s or 100 events | Whichever comes first |
| Experiment traffic lock | Immutable after start | Prevents Simpson's paradox |
| Stats significance threshold | p < 0.05 | Standard α level |
| Minimum sample size | 100 exposures per variant | Before showing results |
