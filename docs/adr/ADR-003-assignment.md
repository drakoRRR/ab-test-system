# ADR-003: Deterministic User Assignment

**Status:** Accepted  
**Date:** 2026-05-16

## Context

A/B testing requires:
1. **Stability** — the same user always lands in the same variant
2. **Uniformity** — a 50/50 (or 70/30) split is genuinely uniform
3. **Independence** — assignment in one experiment must not correlate with another
4. **Statelessness** — no need to persist a user→variant mapping in the database

## Decision

**MurmurHash3** (32-bit) with salt = `userID + ":" + experimentKey`

```go
bucket := murmur3.Sum32([]byte(userID + ":" + experimentKey)) % 10000
```

Bucket `[0, 9999]` → 10 000 equal slots → allows 0.01% precision when configuring traffic_percent.

### Variant Assignment Algorithm

```
Variants: [{key:"control", weight:50}, {key:"treatment", weight:50}]
Traffic:  70%  (the remaining 30% are not in the experiment → see default flag value)

1. Is the user in traffic at all?
   inTraffic = bucket < (trafficPercent * 100)  // bucket < 7000

2. Which variant?
   variantBucket = bucket % totalWeight         // bucket % 100
   walk cumulative weights → return variant key
```

### Why 10 000 slots instead of 100

- 100 slots → 1% precision: impossible to configure a 33.3% rollout
- 10 000 slots → 0.01% precision: sufficient for any real scenario
- Cost: `% 10000` vs `% 100` — zero performance difference

## Why MurmurHash3 (vs MD5, SHA1, FNV)

| Hash | Distribution | Speed | Dependency |
|---|---|---|---|
| MurmurHash3 | Excellent | ~1ns | github.com/spaolacci/murmur3 |
| FNV-1a | Good | ~2ns | stdlib (hash/fnv) |
| SHA-256 | Excellent | ~50ns | stdlib (crypto/sha256) |
| MD5 | Good | ~20ns | stdlib |

MurmurHash3 is the industry standard for feature flag systems (LaunchDarkly, Unleash, and GrowthBook all use it). FNV-1a is an alternative with no external dependency and near-identical results.

**To avoid the external dependency:** swap in `fnv.New32a()` from `hash/fnv` — behavior is identical for our purposes.

## Salt Design

Salt = `userID + ":" + experimentKey`

- **Why not just userID:** different experiments would share the same bucket → correlation → statistical error
- **Why ":" as separator:** prevents collisions between `user="abc", exp="def"` and `user="abcde", exp="f"`

## Assignment Lock

Once an experiment transitions to `running` status:
- `traffic_percent` — immutable
- `variants[].weight` — immutable

**Why this matters:**
Changing traffic from 50% → 70% mid-test would cause some users to switch groups. This violates the independence assumption and makes results statistically invalid (Simpson's Paradox).

**Implementation:** the backend returns 409 Conflict on any attempt to modify these fields while status is `running`.

## Consequences

- Assignment is computed both in the SDK (locally) and in the `/sdk/evaluate` endpoint — **the exact same algorithm in both places**
- Determinism is guaranteed: identical input always produces identical output
- No `assignments` table needed — reduces write load
- If userID is empty (anonymous) — the SDK generates an anonymous ID and stores it in localStorage (for the demo app)
