# SplitLab Go SDK

A Go client library for the SplitLab A/B testing and feature flag platform.
The SDK evaluates flags and experiments **locally** — no network round-trip per user request — and batches analytics events asynchronously in the background.

## Requirements

- Go 1.23+
- A running SplitLab backend
- An SDK API key (created in the SplitLab admin UI under **Project → Settings → SDK Keys**)

## Installation

```bash
go get github.com/drakoRRR/ab-test-system/sdk
```

## Quick Start

```go
package main

import (
    "log"

    "github.com/drakoRRR/ab-test-system/sdk"
)

func main() {
    client, err := sdk.New(sdk.Config{
        APIKey:     "sk_live_...",
        ServiceURL: "https://api.your-splitlab.com/api/v1",
    })
    if err != nil {
        log.Fatalf("sdk init: %v", err)
    }
    defer client.Close()

    // Feature flag check
    if client.IsEnabled("new-checkout", userID) {
        // show new checkout UI
    }

    // A/B experiment assignment
    variant := client.GetVariant("checkout-btn", userID)
    switch variant {
    case "control":
        // show blue button
    case "treatment":
        // show green button
    }

    // Track a conversion
    client.Track("purchase", userID, orderValue)
}
```

## Configuration

```go
client, err := sdk.New(sdk.Config{
    // Required
    APIKey:     "sk_live_...",           // SDK key from SplitLab admin UI
    ServiceURL: "http://localhost:8080/api/v1",

    // Optional — shown with defaults
    PollInterval:   30 * time.Second,   // how often to refresh flag/experiment config
    FlushInterval:  1 * time.Second,    // how often to flush buffered events
    FlushBatchSize: 100,                // flush immediately when buffer reaches this size
    MaxBufferSize:  10_000,             // drop events beyond this (backpressure protection)
    Logger:         zerolog.Nop(),      // structured logger; nop by default
    HTTPClient:     &http.Client{Timeout: 10 * time.Second},
})
```

### Enabling structured logging

```go
import "github.com/rs/zerolog"
import "os"

logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

client, err := sdk.New(sdk.Config{
    APIKey:     "sk_live_...",
    ServiceURL: "https://api.your-splitlab.com/api/v1",
    Logger:     logger,
})
```

The SDK logs at `Warn` for degraded states (stale config, dropped events) and `Debug` for normal operations.

## API Reference

### `sdk.New(cfg Config) (*Client, error)`

Initialises the client. Performs a **synchronous** config fetch before returning — flags and experiments are available to evaluate immediately after `New()` succeeds.

Returns an error if:
- `APIKey` or `ServiceURL` are empty
- The backend is unreachable after 3 retries (with exponential back-off: 500 ms, 1 s, 2 s)

### `client.IsEnabled(flagKey, userID string) bool`

Evaluates a feature flag for the given user. Returns `false` if:
- The flag does not exist
- The flag is disabled
- The user falls outside the percentage rollout
- `userID` is empty

Evaluation is purely local — no network call.

```go
if client.IsEnabled("dark-mode", userID) {
    renderDarkTheme()
}
```

### `client.GetVariant(experimentKey, userID string) string`

Returns the assigned variant key (`"control"`, `"treatment"`, etc.) for the given user.
Returns `""` if the user is outside the traffic allocation or the experiment is not running.

**Side effect:** automatically enqueues an `exposure` event when a non-empty variant is returned. The exposure is attributed to the experiment and variant for conversion rate calculation.

```go
variant := client.GetVariant("checkout-btn", userID)
if variant == "" {
    // user not in experiment — show default UI
    return
}
renderVariant(variant)
```

### `client.Track(eventName, userID string, value float64)`

Enqueues a `conversion` event attributed to all active experiment assignments for `userID`.

- **Non-blocking**: the event is buffered and flushed asynchronously.
- Safe to call at high frequency (e.g., on every API request).
- If no assignment exists for the user, the event is dropped and a warning is logged.

```go
// Track a binary conversion (value = 1.0)
client.Track("purchase", userID, 1.0)

// Track a revenue event
client.Track("revenue", userID, orderTotal)
```

> **Important:** call `GetVariant` before `Track` for the same `userID`. The SDK links conversions to the exposure recorded by `GetVariant`. If `GetVariant` was never called (or returned `""`), there is no assignment to attribute the conversion to.

### `client.Close() error`

Stops background goroutines and flushes all buffered events. **Always call `Close` before your process exits.**

```go
defer func() {
    if err := client.Close(); err != nil {
        log.Printf("sdk close: %v", err)
    }
}()
```

`Close` is idempotent — safe to call multiple times.

## Event Model

The SDK tracks two event types automatically:

| Event | When | Triggered by |
|---|---|---|
| `exposure` | User is assigned to an experiment variant | `GetVariant` (automatic) |
| `conversion` | User completes the goal action | `Track` (explicit) |

Events are batched and sent to `POST /api/v1/sdk/events`. The backend deduplicates by event ID, so retries are safe.

## Deterministic Assignment

User-to-variant assignment is stable across SDK restarts and backend restarts:

```
bucket = MurmurHash3_32(userID + ":" + experimentKey) % 10_000
```

- Same `(userID, experimentKey)` pair always produces the same bucket.
- The SDK and backend use the same algorithm — assignment is consistent regardless of which side evaluates it.
- No database lookup is needed per evaluation.

## Lifecycle in a Web Server

```go
var sdkClient *sdk.Client

func main() {
    var err error
    sdkClient, err = sdk.New(sdk.Config{
        APIKey:     os.Getenv("SDK_API_KEY"),
        ServiceURL: os.Getenv("SDK_SERVICE_URL"),
    })
    if err != nil {
        log.Fatalf("sdk: %v", err)
    }
    defer sdkClient.Close()

    http.HandleFunc("/checkout", handleCheckout)
    http.ListenAndServe(":8080", nil)
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")

    variant := sdkClient.GetVariant("checkout-btn", userID)
    // ... render variant-specific UI ...

    // On successful purchase:
    sdkClient.Track("purchase", userID, orderTotal)
}
```

**Create one `Client` at startup and share it across all request goroutines.** `Client` is safe for concurrent use. Never copy a `Client` value.

## Error Behaviour

The SDK is designed to be safe in degraded states:

| Situation | Behaviour |
|---|---|
| Backend unreachable after initial connect | `New()` returns an error — do not start the application |
| Config refresh fails during polling | Warning logged, evaluation continues with last known config |
| Event buffer full | New events are dropped, warning logged |
| Unknown flag key | `IsEnabled` returns `false` |
| Unknown experiment key | `GetVariant` returns `""` |
| Empty `userID` | All methods return safe defaults, no events enqueued |

## Testing

Inject the SDK via an interface to keep tests fast and hermetic:

```go
type featureClient interface {
    IsEnabled(flagKey, userID string) bool
    GetVariant(experimentKey, userID string) string
    Track(eventName, userID string, value float64)
}

type Handler struct {
    sdk featureClient
}
```

In tests, provide a simple stub:

```go
type stubSDK struct {
    variant string
    enabled bool
}

func (s *stubSDK) IsEnabled(_, _ string) bool  { return s.enabled }
func (s *stubSDK) GetVariant(_, _ string) string { return s.variant }
func (s *stubSDK) Track(_, _ string, _ float64) {}
```

## License

Part of the SplitLab diploma project. See repository root for licence details.
