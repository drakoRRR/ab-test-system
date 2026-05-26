package sdk

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Client is safe for concurrent use. Create once, share globally, never copy.
type Client struct {
	cfg     Config
	cache   *configCache
	buffer  *eventBuffer
	httpCli httpClient

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	assigns sync.Map // map[string]assignment — keyed by userID+":"+experimentID
}

// New initialises the client with a synchronous config fetch. Returns an error if
// APIKey or ServiceURL are empty, or if the backend is unreachable.
func New(cfg Config) (*Client, error) {
	cfg = cfg.withDefaults()
	if cfg.APIKey == "" || cfg.ServiceURL == "" {
		return nil, errors.New("sdk: APIKey and ServiceURL are required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		cfg:     cfg,
		cache:   newConfigCache(),
		buffer:  newEventBuffer(cfg),
		httpCli: newDefaultHTTPClient(cfg.ServiceURL, cfg.HTTPClient),
		cancel:  cancel,
	}

	if err := c.fetchAndStore(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("sdk: initial config fetch: %w", err)
	}

	c.wg.Add(2)
	go c.pollLoop(ctx)
	go c.flushLoop(ctx)

	return c, nil
}

// IsEnabled returns false for unknown flags, disabled flags, and nil config.
func (c *Client) IsEnabled(flagKey, userID string) bool {
	cfg := c.cache.snapshot()
	if cfg == nil {
		return false
	}
	for _, f := range cfg.Flags {
		if f.Key == flagKey {
			return evaluateFlag(f, userID)
		}
	}
	return false
}

// GetVariant returns the variant key for userID, or "" if outside the traffic allocation.
// Tracks an exposure event automatically on every non-empty result.
func (c *Client) GetVariant(experimentKey, userID string) string {
	cfg := c.cache.snapshot()
	if cfg == nil || userID == "" {
		return ""
	}

	for _, exp := range cfg.Experiments {
		if exp.Key != experimentKey {
			continue
		}

		variantKey := evaluateExperiment(exp, userID)
		if variantKey == "" {
			return ""
		}

		var variantID string
		for _, v := range exp.Variants {
			if v.Key == variantKey {
				variantID = v.ID
				break
			}
		}

		c.assigns.Store(userID+":"+exp.ID, assignment{ExperimentID: exp.ID, VariantID: variantID})

		e := newEvent("exposure", userID, exp.ID, variantID, "", 0)
		c.buffer.add(e, c.cfg.Logger)

		return variantKey
	}
	return ""
}

// Track enqueues a conversion event attributed to all active assignments for userID.
func (c *Client) Track(eventName, userID string, value float64) {
	if userID == "" {
		return
	}

	sent := false
	prefix := userID + ":"

	c.assigns.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
			return true
		}
		a := v.(assignment)
		e := newEvent("conversion", userID, a.ExperimentID, a.VariantID, eventName, value)
		c.buffer.add(e, c.cfg.Logger)
		sent = true
		return true
	})

	if !sent {
		c.cfg.Logger.Warn().
			Str("user_id", userID).
			Str("event_name", eventName).
			Msg("sdk: Track called but no assignment found for user — event dropped")
	}
}

// Close stops background goroutines and flushes buffered events. Idempotent.
func (c *Client) Close() error {
	var err error
	c.once.Do(func() {
		c.cancel()
		c.wg.Wait()
	})
	return err
}
