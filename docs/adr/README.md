# Architecture Decision Records

ADRs are short documents that capture important architectural decisions: context, alternatives considered, and rationale.

| ADR | Decision | Status |
|---|---|---|
| [ADR-001](ADR-001-database.md) | PostgreSQL as primary data store (vs Firestore, ClickHouse) | Accepted |
| [ADR-002](ADR-002-event-pipeline.md) | Kafka for async ingestion + idempotent DB deduplication via event_id | Accepted |
| [ADR-003](ADR-003-assignment.md) | MurmurHash3 deterministic user assignment | Accepted |
| [ADR-004](ADR-004-analytics.md) | On-demand analytics + pure Go z-test | Accepted |
| [ADR-005](ADR-005-sdk-config-delivery.md) | SDK config polling (30s TTL) vs SSE | Accepted |
| [ADR-006](ADR-006-multitenancy.md) | Shared DB, row-level isolation by project_id | Accepted |
| [ADR-007](ADR-007-redis.md) | Redis for config cache (cache-aside), Pub/Sub invalidation, rate limiting | Accepted |

## How to Add a New ADR

```
docs/adr/ADR-NNN-short-title.md
```

Structure:
- **Context** — what problem we are solving
- **Decision** — what was decided
- **Rationale** — why (with comparison of alternatives)
- **Consequences** — what changes, trade-offs
