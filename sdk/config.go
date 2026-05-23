package sdk

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	APIKey     string // required — SDK key from SplitLab
	ServiceURL string // required — e.g. "http://localhost:8080/api/v1"

	PollInterval   time.Duration  // default: 30s
	FlushInterval  time.Duration  // default: 1s
	FlushBatchSize int            // default: 100
	MaxBufferSize  int            // default: 10_000
	Logger         zerolog.Logger // default: nop logger
	HTTPClient     *http.Client   // default: 10s timeout
}

func (c Config) withDefaults() Config {
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
	return c
}
