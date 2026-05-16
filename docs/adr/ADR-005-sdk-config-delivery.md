# ADR-005: SDK Config Delivery — Polling Model

**Status:** Accepted  
**Date:** 2026-05-16

## Context

The SDK needs up-to-date flag and experiment configurations to evaluate assignments locally, without a network call per user.

Config delivery options:
1. **Polling** — SDK fetches `/sdk/config` every N seconds
2. **SSE (Server-Sent Events)** — backend pushes updates to all connected SDKs
3. **WebSocket** — bidirectional streaming
4. **Remote evaluation** — every `EvaluateExperiment()` call hits the backend (no local cache)

## Decision

**Polling** with a TTL of **30 seconds**.

## Rationale

| Criterion | Polling | SSE | Remote eval |
|---|---|---|---|
| Implementation | Trivial | Medium | Trivial |
| Config update latency | ~30s | <1s | 0 (always fresh) |
| Stateless backend | Yes | No (connection pool) | Yes |
| Evaluation latency | ~0 (local) | ~0 (local) | ~5–20ms (network) |
| Kill switch activation | ~30s | ~1s | Immediate |

**Why not SSE:**
- Backend becomes stateful — connection pool management required
- On backend restart, all SDKs must reconnect with backoff logic
- A 30-second kill switch delay is acceptable for this project

**Why not remote evaluation:**
- Every `EvaluateExperiment()` becomes a network call → latency in the hot path
- Under a k6 traffic spike, the backend receives N eval requests/s instead of 1 config poll
- The purpose of the SDK is to decouple evaluation from the network

## Config Endpoint

```
GET /sdk/config
X-API-Key: sdk_key_...

Response:
{
  "flags": [
    {
      "key": "new-checkout",
      "enabled": true,
      "rules": [{"type": "percentage", "value": 50}]
    }
  ],
  "experiments": [
    {
      "key": "checkout-btn-color",
      "status": "running",
      "traffic_percent": 70,
      "variants": [
        {"key": "control",   "weight": 50},
        {"key": "treatment", "weight": 50}
      ]
    }
  ],
  "updated_at": "2026-05-16T10:00:00Z"
}
```

## SDK Cache Design

```go
type ConfigCache struct {
    mu        sync.RWMutex
    config    *SDKConfig
    fetchedAt time.Time
    ttl       time.Duration
}

func (c *ConfigCache) Get() *SDKConfig {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.config  // caller reads a snapshot, does not block writes
}

// Background goroutine refreshes the cache
func (c *ConfigCache) refreshLoop(ctx context.Context) {
    ticker := time.NewTicker(c.ttl)
    for {
        select {
        case <-ticker.C:
            cfg, err := c.client.FetchConfig(ctx)
            if err != nil {
                // log + keep serving the stale config
                continue
            }
            c.mu.Lock()
            c.config = cfg
            c.fetchedAt = time.Now()
            c.mu.Unlock()
        case <-ctx.Done():
            return
        }
    }
}
```

**First call:** synchronous fetch during `sdk.Init()` — blocks until config is received. If the backend is unavailable → retry with exponential backoff (3 attempts, then return an error).

## Consequences

- Kill switch takes up to 30s to propagate — acceptable for thesis; in production, reduce TTL to 5s
- Backend receives N config polls / 30s instead of N eval calls / s — significantly lower load
- Stale-while-revalidate: on a fetch error, the SDK continues serving the stale config rather than crashing
- `updated_at` in the response allows the SDK to skip updates when the config has not changed (conditional GET as a future optimization)
