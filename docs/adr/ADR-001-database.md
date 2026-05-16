# ADR-001: PostgreSQL as Primary Data Store

**Status:** Accepted  
**Date:** 2026-05-16

## Context

The system stores several types of data with distinct access patterns:
- Structured entities (projects, flags, experiments) — infrequent writes, frequent reads
- Events (exposure, conversion) — very frequent batch writes, analytical reads (GROUP BY)
- SDK config — frequent reads, infrequent writes

Options considered:
1. **PostgreSQL** — relational, SQL, ACID
2. **Firebase Firestore** — already in use for Auth
3. **PostgreSQL + ClickHouse** — PG for structured data, ClickHouse for events

## Decision

Use **PostgreSQL** for all data.

## Rationale

**Why not Firestore:**
- NoSQL makes analytical queries painful (GROUP BY, window functions, JOINs)
- Event aggregation like `COUNT(*) WHERE experiment_id=X GROUP BY variant_id, event_type` either requires reading all documents (expensive) or manual denormalization
- Statistical computations (z-test feeds on direct COUNT results) require no transformation with SQL

**Why not a separate ClickHouse for events:**
- The project is focused on the platform, not distributed analytics infrastructure
- PostgreSQL handles millions of rows with proper indexes — sufficient for synthetic k6 load
- Upgrade path: if events queries become slow, the `events` table can be partitioned by `ts` or migrated to ClickHouse without touching the rest of the system

**Why PostgreSQL:**
- ACID guarantees — essential for `ON CONFLICT (event_id) DO NOTHING` (deduplication)
- `jsonb` for `flags.rules` — flexible targeting without a separate table
- `date_trunc('hour', ts)` natively — for timeseries queries
- `COPY` / bulk insert — efficient batch writes for events
- Single service — simpler Docker Compose, simpler migrations

## Consequences

- An indexing strategy for `events` is required (see architecture.md)
- At >50M events/month, consider range partitioning by `ts`
- Upgrade path to ClickHouse remains open — the service layer isolates storage details
