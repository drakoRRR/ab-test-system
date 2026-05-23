package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type configCache struct {
	mu     sync.RWMutex
	config *SDKConfig
}

func (c *configCache) snapshot() *SDKConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

func (c *configCache) store(cfg *SDKConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cfg
}

func newConfigCache() *configCache {
	return &configCache{
		config: nil,
	}
}

// fetchAndStore retries up to 3 times with exponential backoff (500ms, 1s, 2s).
func (c *Client) fetchAndStore(ctx context.Context) error {
	var err error
	for attempt := range 3 {
		cfg, fetchErr := c.httpCli.FetchConfig(ctx, c.cfg.APIKey)
		if fetchErr == nil {
			c.cache.store(cfg)
			return nil
		}
		err = fetchErr
		wait := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond // 500ms, 1s, 2s
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("after 3 attempts: %w", err)
}

// pollLoop refreshes config every PollInterval. Stale-while-revalidate: fetch errors
// keep the last known config rather than breaking evaluation.
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
				continue
			}
			c.cache.store(cfg)
		case <-ctx.Done():
			return
		}
	}
}
