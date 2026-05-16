# Go SDK Rules

Rules for all code under `sdk/`.
The SDK is a Go library embedded in applications (like `demo/`) to evaluate feature flags and experiments locally, and track events asynchronously.

---

## Module Structure

```
sdk/
  go.mod              # module: github.com/drakoRRR/ab-test-system/sdk
  sdk.go              # public API: New(), Client, IsEnabled(), GetVariant(), Track(), Close()
  config.go           # Config struct with documented defaults
  types.go            # internal types: SDKConfig, Experiment, Flag, Variant, Event
  evaluate.go         # pure evaluation functions — no I/O, no locks, no goroutines
  poller.go           # ConfigCache + background refresh goroutine
  buffer.go           # EventBuffer: channel, flush goroutine, UUID generation
  http.go             # HTTPClient interface + default implementation
  evaluate_test.go
  buffer_test.go
  sdk_test.go
```

`go.mod` is a **separate module** from the backend. It depends only on:
- `github.com/spaolacci/murmur3` — deterministic hashing
- `github.com/google/uuid` — event ID generation
- `github.com/rs/zerolog` — structured logging

No other external dependencies. If a dependency can be replaced with stdlib, replace it.

---

## Public API

The public surface is small and intentional. Do not add methods without a clear use case.

```go
// sdk/sdk.go

// New initialises the client. It performs a synchronous config fetch before
// returning — the caller can evaluate flags immediately after New() succeeds.
// Returns an error if the backend is unreachable after retries.
func New(cfg Config) (*Client, error)

// Client is safe for concurrent use by multiple goroutines.
// Create once, share globally within the application. Never copy.
type Client struct { ... }

// IsEnabled reports whether the flag is enabled for userID.
// Returns false if the flag does not exist, the config is unavailable, or
// userID is empty. Never panics.
func (c *Client) IsEnabled(flagKey, userID string) bool

// GetVariant returns the assigned variant key for userID in the experiment.
// Returns "" if the experiment is not running, the user is outside the
// traffic allocation, or the config is unavailable.
// When a non-empty variant is returned the exposure event is tracked automatically.
func (c *Client) GetVariant(experimentKey, userID string) string

// Track enqueues a conversion event. It is non-blocking — the event is
// buffered and flushed asynchronously. Safe to call at high frequency.
func (c *Client) Track(eventName, userID string, value float64)

// Close flushes all buffered events, stops background goroutines, and
// releases resources. Idempotent — safe to call multiple times.
// The client must not be used after Close.
func (c *Client) Close() error
```

### Safe defaults over errors

Public methods **never panic** and **never return errors** for evaluation.
When the config is stale or the key does not exist, return the safe default:

| Method | Safe default | Reason |
|---|---|---|
| `IsEnabled` | `false` | unknown flag = disabled |
| `GetVariant` | `""` | unknown experiment = not enrolled |
| `Track` | no-op | drop event, log warning |

Only `New()` and `Close()` return errors — they are lifecycle operations.

---

## Config

```go
// sdk/config.go

type Config struct {
    APIKey     string        // required — SDK key from the platform
    ServiceURL string        // required — e.g. "https://api.example.com"

    PollInterval   time.Duration // default: 30s
    FlushInterval  time.Duration // default: 1s
    FlushBatchSize int           // default: 100; flush when buffer reaches this size
    MaxBufferSize  int           // default: 10_000; drop events beyond this

    Logger     zerolog.Logger // optional; defaults to zerolog.Nop()
    HTTPClient *http.Client   // optional; defaults to &http.Client{Timeout: 10s}
}

func (c *Config) withDefaults() Config {
    if c.PollInterval == 0 {
        c.PollInterval = 30 * time.Second
    }
    if c.FlushInterval == 0 {
        c.FlushInterval = time.Second
    }
    if c.FlushBatchSize == 0 {
        c.FlushBatchSize = 100
    }
    if c.MaxBufferSize == 0 {
        c.MaxBufferSize = 10_000
    }
    if c.HTTPClient == nil {
        c.HTTPClient = &http.Client{Timeout: 10 * time.Second}
    }
    return *c
}
```

Validate `APIKey` and `ServiceURL` in `New()` — return an error immediately if either is empty.

---

## Lifecycle: goroutines and shutdown

All background goroutines are started in `New()` and stopped in `Close()`.
Use a single `context.Context` derived from `context.Background()` to control all goroutines.

```go
type Client struct {
    cfg     Config
    cache   *configCache
    buffer  *eventBuffer
    httpCli httpClient

    cancel context.CancelFunc
    wg     sync.WaitGroup
    once   sync.Once // guards Close()
}

func New(cfg Config) (*Client, error) {
    cfg = cfg.withDefaults()
    if cfg.APIKey == "" || cfg.ServiceURL == "" {
        return nil, errors.New("sdk: APIKey and ServiceURL are required")
    }

    ctx, cancel := context.WithCancel(context.Background())

    c := &Client{
        cfg:    cfg,
        cache:  newConfigCache(),
        buffer: newEventBuffer(cfg),
        cancel: cancel,
    }

    // Synchronous first fetch — caller can evaluate immediately after New()
    if err := c.fetchAndStore(ctx); err != nil {
        cancel()
        return nil, fmt.Errorf("sdk: initial config fetch failed: %w", err)
    }

    c.wg.Add(2)
    go c.pollLoop(ctx)   // refreshes config every PollInterval
    go c.flushLoop(ctx)  // flushes event buffer every FlushInterval

    return c, nil
}

func (c *Client) Close() error {
    var err error
    c.once.Do(func() {
        c.cancel()      // signal all goroutines to stop
        c.wg.Wait()     // wait for clean exit
        err = c.flush(context.Background()) // final flush of remaining events
    })
    return err
}
```

**Rules:**
- Every `go func()` launched in `New()` must call `c.wg.Done()` on exit
- `c.wg.Wait()` in `Close()` ensures no goroutine is mid-flight when we flush
- `sync.Once` makes `Close()` idempotent

---

## Evaluation Engine

Evaluation is a **pure function** in `evaluate.go` — no I/O, no locks, no side effects.
It takes a config snapshot and returns a result synchronously.

```go
// sdk/evaluate.go

// assignBucket returns a deterministic bucket [0, 9999] for the user+key pair.
// This is the core of the assignment algorithm — must be identical to the backend.
func assignBucket(userID, key string) uint32 {
    return murmur3.Sum32([]byte(userID + ":" + key)) % 10000
}

// evaluateFlag returns true if the flag is enabled for userID.
func evaluateFlag(flag Flag, userID string) bool {
    if !flag.Enabled {
        return false
    }
    for _, rule := range flag.Rules {
        if rule.Type == "percentage" {
            bucket := assignBucket(userID, flag.Key)
            return bucket < uint32(rule.Value*100)
        }
    }
    return flag.Enabled
}

// evaluateExperiment returns the variant key assigned to userID, or "".
func evaluateExperiment(exp Experiment, userID string) string {
    if exp.Status != "running" {
        return ""
    }

    bucket := assignBucket(userID, exp.Key)

    // Step 1: is the user in the traffic allocation?
    if bucket >= uint32(exp.TrafficPercent*100) {
        return ""
    }

    // Step 2: which variant?
    totalWeight := 0
    for _, v := range exp.Variants {
        totalWeight += v.Weight
    }

    variantBucket := int(bucket) % totalWeight
    cumulative := 0
    for _, v := range exp.Variants {
        cumulative += v.Weight
        if variantBucket < cumulative {
            return v.Key
        }
    }

    return ""
}
```

**Critical invariant:** `assignBucket` must produce **exactly the same output** as the backend's assignment endpoint for identical inputs. Any change here requires a matching change on the backend.

Tests must verify this with fixed known inputs:

```go
func TestAssignBucket_KnownValues(t *testing.T) {
    tests := []struct {
        userID string
        key    string
        want   uint32
    }{
        {"user-123", "checkout-btn", 4821},   // record actual value and pin it
        {"user-456", "checkout-btn", 7103},
        {"user-123", "banner-color", 2947},   // same user, different key → different bucket
    }
    for _, tt := range tests {
        t.Run(tt.userID+"/"+tt.key, func(t *testing.T) {
            require.Equal(t, tt.want, assignBucket(tt.userID, tt.key))
        })
    }
}
```

---

## Config Cache and Polling

Exactly **one goroutine** writes to the cache — the poller. Readers use `sync.RWMutex`.

```go
// sdk/poller.go

type configCache struct {
    mu     sync.RWMutex
    config *SDKConfig
}

// snapshot returns the current config for reading. Callers must not mutate it.
func (c *configCache) snapshot() *SDKConfig {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.config
}

func (c *configCache) store(cfg *SDKConfig) {
    c.mu.Lock()
    c.config = cfg
    c.mu.Unlock()
}

func (c *Client) pollLoop(ctx context.Context) {
    defer c.wg.Done()
    ticker := time.NewTicker(c.cfg.PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            cfg, err := c.httpCli.FetchConfig(ctx, c.cfg.APIKey)
            if err != nil {
                c.cfg.Logger.Warn().Err(err).Msg("sdk: config refresh failed, using stale config")
                continue  // stale-while-revalidate: keep serving old config
            }
            c.cache.store(cfg)
        case <-ctx.Done():
            return
        }
    }
}
```

**Stale-while-revalidate is mandatory.** A fetch error must never crash the SDK or block evaluation — log a warning and continue with the last known config.

Initial fetch in `New()` uses exponential backoff:

```go
func (c *Client) fetchAndStore(ctx context.Context) error {
    var err error
    for attempt := range 3 {
        cfg, fetchErr := c.httpCli.FetchConfig(ctx, c.cfg.APIKey)
        if fetchErr == nil {
            c.cache.store(cfg)
            return nil
        }
        err = fetchErr
        wait := time.Duration(1<<attempt) * 500 * time.Millisecond // 500ms, 1s, 2s
        select {
        case <-time.After(wait):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return fmt.Errorf("after 3 attempts: %w", err)
}
```

---

## Event Buffer

The buffer is a **buffered channel**. `Add` is non-blocking — if the channel is full, the event is dropped with a warning (backpressure protection).

```go
// sdk/buffer.go

type event struct {
    ID           string    // uuid.NewString() — generated at creation, used for deduplication
    ProjectID    string
    UserID       string
    ExperimentID string
    VariantID    string
    Type         string    // "exposure" | "conversion"
    Name         string
    Value        float64
    Timestamp    time.Time
}

type eventBuffer struct {
    ch  chan event
    cfg Config
}

func newEventBuffer(cfg Config) *eventBuffer {
    return &eventBuffer{
        ch:  make(chan event, cfg.MaxBufferSize),
        cfg: cfg,
    }
}

// add enqueues an event. Non-blocking — drops and logs if the buffer is full.
func (b *eventBuffer) add(e event, log zerolog.Logger) {
    select {
    case b.ch <- e:
    default:
        log.Warn().Str("event_type", e.Type).Msg("sdk: event buffer full, dropping event")
    }
}

func (c *Client) flushLoop(ctx context.Context) {
    defer c.wg.Done()
    ticker := time.NewTicker(c.cfg.FlushInterval)
    defer ticker.Stop()

    batch := make([]event, 0, c.cfg.FlushBatchSize)

    for {
        select {
        case e := <-c.buffer.ch:
            batch = append(batch, e)
            if len(batch) >= c.cfg.FlushBatchSize {
                c.sendBatch(ctx, batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            if len(batch) > 0 {
                c.sendBatch(ctx, batch)
                batch = batch[:0]
            }

        case <-ctx.Done():
            // drain remaining events before exit
            for len(c.buffer.ch) > 0 {
                batch = append(batch, <-c.buffer.ch)
            }
            if len(batch) > 0 {
                c.sendBatch(context.Background(), batch)
            }
            return
        }
    }
}
```

**Event ID is generated before buffering**, not before sending. The same ID is preserved on retry, which is what enables deduplication on the backend (`ON CONFLICT (id) DO NOTHING`).

```go
func newEvent(typ, userID, experimentID, variantID, name string, value float64) event {
    return event{
        ID:           uuid.NewString(), // generated once here
        Type:         typ,
        UserID:       userID,
        ExperimentID: experimentID,
        VariantID:    variantID,
        Name:         name,
        Value:        value,
        Timestamp:    time.Now().UTC(),
    }
}
```

---

## HTTP Client Interface

The HTTP client is hidden behind an interface so tests can replace it without a real server.

```go
// sdk/http.go

type httpClient interface {
    FetchConfig(ctx context.Context, apiKey string) (*SDKConfig, error)
    SendEvents(ctx context.Context, apiKey string, batch []event) error
}

// defaultHTTPClient is the real implementation used in production.
type defaultHTTPClient struct {
    baseURL string
    client  *http.Client
}
```

Tests inject a `mockHTTPClient` that returns pre-configured responses.
Never use `net/http/httptest` at the unit test level — reserve it for integration tests.

---

## Testing

### Unit tests — pure evaluation

Test `evaluate.go` exhaustively with table-driven tests and zero mocks.
These are the most important tests in the SDK — the evaluation algorithm must be correct.

```go
func TestEvaluateExperiment(t *testing.T) {
    running := Experiment{
        Key:            "btn-color",
        Status:         "running",
        TrafficPercent: 70,
        Variants: []Variant{
            {Key: "control",   Weight: 50},
            {Key: "treatment", Weight: 50},
        },
    }

    tests := []struct {
        name   string
        exp    Experiment
        userID string
        want   string
    }{
        {
            name:   "assigns variant for user in traffic",
            exp:    running,
            userID: "user-in-traffic",
            want:   "control",           // pin actual hash result
        },
        {
            name:   "excludes user outside traffic",
            exp:    running,
            userID: "user-out-of-traffic",
            want:   "",
        },
        {
            name:   "returns empty for paused experiment",
            exp:    Experiment{Key: "btn-color", Status: "paused"},
            userID: "user-123",
            want:   "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            require.Equal(t, tt.want, evaluateExperiment(tt.exp, tt.userID))
        })
    }
}
```

### Unit tests — buffer

Test flush triggers: size threshold and time threshold.
Use a `mockHTTPClient` and verify `SendEvents` call count and contents.

### Integration tests

`sdk_test.go` tests the full `New() → evaluate → track → Close()` lifecycle
using a `mockHTTPClient` — no real HTTP server needed at this level.

```go
func TestClient_GetVariant_TracksExposure(t *testing.T) {
    mock := &mockHTTPClient{
        config: &SDKConfig{
            Experiments: []Experiment{{
                Key: "my-exp", Status: "running",
                TrafficPercent: 100,
                Variants: []Variant{{Key: "control", Weight: 100}},
            }},
        },
    }

    c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
    require.NoError(t, err)
    defer c.Close()

    variant := c.GetVariant("my-exp", "user-1")
    require.Equal(t, "control", variant)

    c.Close()
    require.Len(t, mock.sentEvents, 1)
    require.Equal(t, "exposure", mock.sentEvents[0].Type)
}
```

---

## Concurrency Rules

- `configCache` is the only shared mutable state — protected by `sync.RWMutex`
- `eventBuffer.ch` is the only channel — never close it explicitly; let `ctx.Done()` drain it
- No other package-level or struct-level mutexes
- Evaluation (`IsEnabled`, `GetVariant`) calls `cache.snapshot()` — a single pointer read under `RLock`; it is fast enough to call from hot paths
- `Track()` calls `buffer.add()` — a non-blocking channel send; safe to call from hot paths

---

## Miscellaneous

- No `init()` functions — no side effects at package init time
- No global `var client *Client` — the caller owns the lifecycle
- Log with `zerolog` at `Warn` for dropped events or stale config; `Debug` for normal operations
- Keep the `sdk/` module at the minimum viable Go version that the demo app requires
- All exported symbols have a one-line doc comment — the SDK is a public library
