# ADR-004: Analytics — On-Demand Computation, Pure Go Statistics

**Status:** Accepted  
**Date:** 2026-05-16

## Context

After collecting events, the system must compute experiment results:
- Conversion rate per variant
- Uplift (relative difference between treatment and control)
- Statistical significance (p-value, confidence interval)
- Timeseries (how metrics change over time)

Computation options:
1. **On-demand** — read events from DB and compute on every `/analytics` request
2. **Pre-aggregated** — background job writes hourly snapshots into `experiment_stats`
3. **Hybrid** — snapshot + real-time delta

Statistics implementation options:
1. **Pure Go** — implement formulas directly
2. **CGo + GNU Scientific Library** — C library via CGo
3. **HTTP call to a Python/R service** — separate stats microservice

## Decision

**On-demand** computation with a **pure Go** statistical module.

## Rationale

### On-demand vs Pre-aggregation

**On-demand chosen because:**
- At thesis scale (synthetic data, ~100K–1M events) SQL `GROUP BY` runs in <50ms with proper indexes
- Background jobs add complexity: scheduler, idempotency, restart behavior
- On-demand always returns fresh data (no lag)

**When to switch to pre-aggregation:**
- `events` table exceeds 100M rows
- Analytics endpoint receives >100 RPS
- Complex multi-experiment reports are needed

**Upgrade path:** add an `experiment_hourly_stats` table + a cron job — the service layer remains unchanged.

### Pure Go vs External Stats Engine

**Pure Go chosen because:**
- Two tests are sufficient for the thesis: z-test for proportions + Wilson score CI
- Formulas are simple — 20–30 lines of Go
- No CGo dependency (cross-compilation, Docker image size)
- A separate Python service means another container, HTTP latency, and an additional failure mode

## Statistical Methods

### Two-Proportion Z-test

Tests H₀: p_control = p_treatment

```
p1 = conversions_control / exposures_control
p2 = conversions_treatment / exposures_treatment
p_pool = (conversions_control + conversions_treatment) / (exposures_control + exposures_treatment)

SE = sqrt(p_pool * (1 - p_pool) * (1/n1 + 1/n2))
z  = (p2 - p1) / SE
p_value = 2 * (1 - Φ(|z|))   // two-tailed
```

### Confidence Interval (Wilson Score)

Instead of the naive `p ± 1.96*sqrt(p*(1-p)/n)`, we use the Wilson score interval — it is correct for small n and extreme values of p.

```
center = (p + z²/(2n)) / (1 + z²/n)
margin = z * sqrt(p*(1-p)/n + z²/(4n²)) / (1 + z²/n)
CI = [center - margin, center + margin]    // z=1.96 for 95%
```

### Uplift

```
uplift = (p_treatment - p_control) / p_control * 100%
```

### Minimum Sample Size Warning

Before displaying results, we check:
```
if exposures < 100 per variant → return warning "Insufficient data"
```

## SQL Queries

### Aggregation for analytics

```sql
SELECT
  v.key         AS variant_key,
  v.name        AS variant_name,
  COUNT(DISTINCT CASE WHEN e.event_type = 'exposure'   THEN e.user_id END) AS exposures,
  COUNT(DISTINCT CASE WHEN e.event_type = 'conversion' THEN e.user_id END) AS conversions
FROM variants v
LEFT JOIN events e ON e.variant_id = v.id
  AND e.experiment_id = $1
WHERE v.experiment_id = $1
GROUP BY v.id, v.key, v.name;
```

Note: `COUNT(DISTINCT user_id)` counts unique users, not raw event rows.
A single user may produce multiple exposure events (SDK retry) — `DISTINCT` handles deduplication.

### Timeseries

```sql
SELECT
  date_trunc('hour', e.ts) AS hour,
  v.key                    AS variant_key,
  COUNT(DISTINCT CASE WHEN e.event_type = 'exposure'   THEN e.user_id END) AS exposures,
  COUNT(DISTINCT CASE WHEN e.event_type = 'conversion' THEN e.user_id END) AS conversions
FROM events e
JOIN variants v ON v.id = e.variant_id
WHERE e.experiment_id = $1
  AND e.ts BETWEEN $2 AND $3
GROUP BY 1, 2
ORDER BY 1;
```

## Required DB Indexes

```sql
-- Primary analytics query
CREATE INDEX events_experiment_type_variant
  ON events(experiment_id, event_type, variant_id);

-- Timeseries
CREATE INDEX events_experiment_ts
  ON events(experiment_id, ts);

-- Deduplication (covered by PRIMARY KEY, but explicit for clarity)
CREATE UNIQUE INDEX events_id_unique ON events(id);
```

## Consequences

- The first analytics request may be slow without a warm buffer pool — acceptable for thesis scale
- The stats module is pure functions with no state → easy to unit test with known inputs
- On-demand responses are always consistent (reads within a single transaction snapshot)
