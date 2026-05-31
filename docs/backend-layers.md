# Backend Layered Architecture

The `backend/internal/` directory enforces a strict separation of concerns across four layers:

```
backend/internal/
├── domain/     — pure domain types with no external dependencies
├── handler/    — HTTP handlers generated from the OpenAPI specification
├── services/   — business logic (flag management, experiments, analytics computation)
└── infra/      — data access implementations for PostgreSQL, Kafka, Redis, Firebase Auth
```

The `backend/api/` directory contains modular OpenAPI YAML files and a Redocly configuration that bundles them into a single specification document.

---

## Dependency Direction

```
handler  →  services  →  domain
infra    →  domain
cmd      →  handler, services, infra   (wiring only)
```

No inner layer imports an outer one — all dependencies point inward toward `domain`.
