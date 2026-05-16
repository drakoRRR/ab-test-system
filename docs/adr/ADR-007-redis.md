# ADR-007: Redis for Config Cache, Pub/Sub Invalidation, and Rate Limiting

**Status:** Accepted  
**Date:** 2026-05-16

## Context

Three independent problems benefit from Redis:

1. **Config read amplification** — `/sdk/config` is polled by every SDK instance every 30s.
   Each poll currently hits PostgreSQL to query active flags and experiments.
   Adding a cache layer reduces PG load and improves response time.

2. **Kill switch latency** — With TTL-based expiry alone, a flag disable takes up to 30s to
   propagate to the SDK. In production, operators expect a kill switch to take effect faster.
   Without a notification mechanism, the only options are a shorter TTL (more PG load) or
   server-push (SSE/WebSocket complexity).

3. **Event ingestion rate limiting** — The `/sdk/events` endpoint is public (authenticated only
   by SDK key). A misbehaving or compromised SDK key could flood Kafka. Per-key rate limiting
   protects the pipeline. Implementing this correctly requires atomic read-modify-write, which
   is not safely achievable in a single PostgreSQL query without transactions.

## Decision

Add **Redis** with three specific responsibilities:

| Responsibility | Pattern | Redis data structure |
|---|---|---|
| SDK config cache | Cache-aside | `STRING` with TTL |
| Cache invalidation on flag/experiment change | Pub/Sub | `PUBLISH` / `SUBSCRIBE` |
| Per-SDK-key rate limiting on `/sdk/events` | Sliding window | `SORTED SET` + Lua |

**Go client:** `github.com/redis/go-redis/v9` — context-aware, idiomatic API.

---

## Pattern 1: Config Cache (Cache-Aside)

```
GET /sdk/config (SDK poll)
        │
        ▼
  Redis GET "config:{projectID}"
        │
   ┌────┴────┐
 HIT │         │ MISS
        │         │
        │         ▼
        │    SELECT flags + experiments FROM pg
        │         │
        │         ▼
        │    Redis SET "config:{projectID}" ttl=30s
        │         │
        └────┬────┘
             │
             ▼
      return config JSON
```

```go
// infra/redis/config_cache.go

func (c *ConfigCache) Get(ctx context.Context, projectID uuid.UUID) (*sdk.Config, error) {
    key := fmt.Sprintf("config:%s", projectID)

    data, err := c.rdb.Get(ctx, key).Bytes()
    if errors.Is(err, redis.Nil) {
        return nil, nil // cache miss — caller fetches from PG
    }
    if err != nil {
        return nil, fmt.Errorf("redis get config: %w", err)
    }

    var cfg sdk.Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }
    return &cfg, nil
}

func (c *ConfigCache) Set(ctx context.Context, projectID uuid.UUID, cfg *sdk.Config) error {
    key := fmt.Sprintf("config:%s", projectID)
    data, err := json.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("marshal config: %w", err)
    }
    return c.rdb.Set(ctx, key, data, 30*time.Second).Err()
}

func (c *ConfigCache) Invalidate(ctx context.Context, projectID uuid.UUID) error {
    key := fmt.Sprintf("config:%s", projectID)
    return c.rdb.Del(ctx, key).Err()
}
```

The service layer orchestrates cache-aside:

```go
// services/sdk/service.go

func (s *Service) GetConfig(ctx context.Context, projectID uuid.UUID) (*sdk.Config, error) {
    // 1. try cache
    cfg, err := s.cache.Get(ctx, projectID)
    if err != nil {
        s.log.Warn().Err(err).Msg("config cache read failed, falling back to DB")
    }
    if cfg != nil {
        return cfg, nil // cache hit
    }

    // 2. cache miss — query PG
    cfg, err = s.repo.BuildConfig(ctx, projectID)
    if err != nil {
        return nil, fmt.Errorf("GetConfig: %w", err)
    }

    // 3. populate cache (best-effort — don't fail the request on cache write error)
    if err := s.cache.Set(ctx, projectID, cfg); err != nil {
        s.log.Warn().Err(err).Msg("config cache write failed")
    }

    return cfg, nil
}
```

---

## Pattern 2: Pub/Sub Cache Invalidation

When a flag or experiment is updated, the service publishes to a Redis channel.
A subscriber goroutine in the same backend process receives the message and deletes the cache key.
The next SDK poll hits PostgreSQL and repopulates the cache with fresh data.

```
UpdateFlag API call
      │
      ▼
  PG UPDATE flags SET ...
      │
      ▼
  Redis PUBLISH "config:invalidated" "{projectID}"
      │                                    │
      │                          (async, separate goroutine)
      │                                    │
      ▼                                    ▼
  200 OK to caller          Redis DEL "config:{projectID}"
                                          │
                                          ▼
                             next /sdk/config → cache miss → fresh PG read
```

```go
// infra/redis/invalidator.go

const invalidationChannel = "config:invalidated"

func (p *ConfigPublisher) Publish(ctx context.Context, projectID uuid.UUID) error {
    return p.rdb.Publish(ctx, invalidationChannel, projectID.String()).Err()
}

// InvalidationSubscriber runs as a goroutine for the lifetime of the process.
type InvalidationSubscriber struct {
    rdb   *redis.Client
    cache *ConfigCache
    log   zerolog.Logger
}

func (s *InvalidationSubscriber) Run(ctx context.Context) {
    sub := s.rdb.Subscribe(ctx, invalidationChannel)
    defer sub.Close()

    ch := sub.Channel()
    for {
        select {
        case msg, ok := <-ch:
            if !ok {
                return
            }
            projectID, err := uuid.Parse(msg.Payload)
            if err != nil {
                s.log.Warn().Str("payload", msg.Payload).Msg("invalid projectID in invalidation message")
                continue
            }
            if err := s.cache.Invalidate(ctx, projectID); err != nil {
                s.log.Warn().Err(err).Msg("cache invalidation failed")
            }

        case <-ctx.Done():
            return
        }
    }
}
```

The service that modifies flags/experiments calls `publisher.Publish` after a successful PG write:

```go
func (s *FlagService) UpdateFlag(ctx context.Context, flagID uuid.UUID, req domain.UpdateFlagRequest) error {
    flag, err := s.repo.Update(ctx, flagID, req)
    if err != nil {
        return fmt.Errorf("UpdateFlag: %w", err)
    }

    // best-effort publish — don't fail the API call if Redis is down
    if err := s.publisher.Publish(ctx, flag.ProjectID); err != nil {
        s.log.Warn().Err(err).Msg("cache invalidation publish failed")
    }

    return nil
}
```

**Why this is better than a shorter TTL:**
A TTL of 5s would guarantee 5s max staleness but costs 12× more PG queries per SDK instance.
Pub/Sub gives near-instant propagation (~100ms) while keeping the 30s TTL as a safety net.

---

## Pattern 3: Sliding Window Rate Limiting

Implemented as a Lua script to ensure atomicity. The script counts events in the current 1-second window using a sorted set.

```lua
-- rate_limit.lua
-- KEYS[1] = "ratelimit:{sdkKey}"
-- ARGV[1] = current timestamp (ms)
-- ARGV[2] = window size (ms) = 1000
-- ARGV[3] = max requests per window

local key      = KEYS[1]
local now      = tonumber(ARGV[1])
local window   = tonumber(ARGV[2])
local limit    = tonumber(ARGV[3])
local min_time = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', min_time)
local count = redis.call('ZCARD', key)

if count >= limit then
    return 0  -- rejected
end

redis.call('ZADD', key, now, now .. '-' .. math.random())
redis.call('PEXPIRE', key, window)
return 1  -- allowed
```

```go
// infra/redis/rate_limiter.go

type RateLimiter struct {
    rdb    *redis.Client
    script *redis.Script
    limit  int
}

func NewRateLimiter(rdb *redis.Client, requestsPerSecond int) *RateLimiter {
    return &RateLimiter{
        rdb:    rdb,
        script: redis.NewScript(rateLimitScript),
        limit:  requestsPerSecond,
    }
}

// Allow returns true if the request is within the rate limit for sdkKey.
func (r *RateLimiter) Allow(ctx context.Context, sdkKey string) (bool, error) {
    key := "ratelimit:" + sdkKey
    now := time.Now().UnixMilli()

    result, err := r.script.Run(ctx, r.rdb, []string{key},
        now, 1000, r.limit,
    ).Int()
    if err != nil {
        // if Redis is down, allow the request — don't block on infra failure
        return true, fmt.Errorf("rate limiter: %w", err)
    }
    return result == 1, nil
}
```

Usage in the events handler:

```go
func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    sdkKey := r.Header.Get("X-API-Key")

    allowed, err := h.rateLimiter.Allow(r.Context(), sdkKey)
    if err != nil {
        h.log.Warn().Err(err).Msg("rate limiter unavailable, allowing request")
    }
    if !allowed {
        writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
        return
    }
    // ... rest of handler
}
```

**Default limit:** 1 000 events/second per SDK key. Adjust per use case.

---

## Failure Modes

Redis is a **supporting store** — its failure must never break core functionality.

| Failure | Behaviour |
|---|---|
| Redis down during config read | Log warning, fall through to PG — config still served |
| Redis down during cache write | Log warning, continue — next request goes to PG again |
| Redis down during Pub/Sub publish | Log warning — cache expires naturally via TTL (30s max staleness) |
| Redis down during rate limiting | Log warning, allow the request — prefer availability over protection |

This is the **fail-open** strategy: Redis degradation downgrades the system to the pre-Redis behaviour, not an outage.

## Docker Compose

```yaml
redis:
  image: redis:7.2-alpine
  ports:
    - "6379:6379"
  command: redis-server --save "" --appendonly no  # no persistence for dev
```

`redis-cli monitor` for debugging during development.

## Consequences

- `+1` Docker Compose service (~30MB RAM for dev workload)
- New `infra/redis/` sub-package; services depend on it via interfaces (same pattern as PG)
- Rate limiter, cache, and publisher are injected into services — no global Redis client
- Redis client is initialised in `cmd/server/main.go` alongside PG and Kafka
- Integration tests for Redis use `testcontainers-go` (same as PG and Kafka)
