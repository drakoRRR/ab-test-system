package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newWithHTTPClient bypasses the real HTTP call in New() for testing.
func newWithHTTPClient(cfg Config, cli httpClient) (*Client, error) {
	cfg = cfg.withDefaults()
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		cfg:     cfg,
		cache:   &configCache{},
		buffer:  newEventBuffer(cfg),
		httpCli: cli,
		cancel:  cancel,
	}
	cfgData, err := cli.FetchConfig(ctx, cfg.APIKey)
	if err != nil {
		cancel()
		return nil, err
	}
	c.cache.store(cfgData)

	c.wg.Add(2)
	go c.pollLoop(ctx)
	go c.flushLoop(ctx)
	return c, nil
}

func TestClient_IsEnabled_FlagEnabled(t *testing.T) {
	mock := &mockHTTPClient{
		configResp: &SDKConfig{
			Flags: []Flag{
				{Key: "my-flag", Enabled: true, Rules: nil},
			},
		},
	}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	require.NoError(t, err)
	defer c.Close()

	require.True(t, c.IsEnabled("my-flag", "any-user"))
}

func TestClient_IsEnabled_FlagDisabled(t *testing.T) {
	mock := &mockHTTPClient{
		configResp: &SDKConfig{
			Flags: []Flag{
				{Key: "my-flag", Enabled: false},
			},
		},
	}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	require.NoError(t, err)
	defer c.Close()

	require.False(t, c.IsEnabled("my-flag", "any-user"))
}

func TestClient_IsEnabled_UnknownFlag_ReturnsFalse(t *testing.T) {
	mock := &mockHTTPClient{configResp: &SDKConfig{}}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	require.NoError(t, err)
	defer c.Close()

	require.False(t, c.IsEnabled("nonexistent", "user-1"))
}

func TestClient_GetVariant_ReturnsVariant(t *testing.T) {
	mock := &mockHTTPClient{
		configResp: &SDKConfig{
			Experiments: []Experiment{
				{
					ID:             "exp-1",
					Key:            "my-exp",
					TrafficPercent: 100,
					Variants: []Variant{
						{ID: "v1", Key: "control", Weight: 100},
					},
				},
			},
		},
	}
	c, err := newWithHTTPClient(Config{
		APIKey: "k", ServiceURL: "http://x",
		FlushInterval: 100 * time.Millisecond,
	}, mock)
	require.NoError(t, err)

	variant := c.GetVariant("my-exp", "user-1")
	require.Equal(t, "control", variant)

	require.NoError(t, c.Close())
	events := mock.events()
	require.Len(t, events, 1)
	require.Equal(t, "exposure", events[0].Type)
	require.Equal(t, "exp-1", events[0].ExperimentID)
	require.Equal(t, "v1", events[0].VariantID)
}

func TestClient_GetVariant_UnknownExperiment_ReturnsEmpty(t *testing.T) {
	mock := &mockHTTPClient{configResp: &SDKConfig{}}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	require.NoError(t, err)
	defer c.Close()

	require.Empty(t, c.GetVariant("nonexistent", "user-1"))
}

func TestClient_Track_SendsConversionEvent(t *testing.T) {
	mock := &mockHTTPClient{
		configResp: &SDKConfig{
			Experiments: []Experiment{
				{
					ID:             "exp-1",
					Key:            "my-exp",
					TrafficPercent: 100,
					Variants: []Variant{
						{ID: "v1", Key: "control", Weight: 100},
					},
				},
			},
		},
	}
	c, err := newWithHTTPClient(Config{
		APIKey: "k", ServiceURL: "http://x",
		FlushInterval: 100 * time.Millisecond,
	}, mock)
	require.NoError(t, err)

	c.GetVariant("my-exp", "user-1")
	c.Track("purchase", "user-1", 99.0)

	require.NoError(t, c.Close())

	events := mock.events()
	require.GreaterOrEqual(t, len(events), 2)

	types := map[string]int{}
	for _, e := range events {
		types[e.Type]++
	}
	require.Equal(t, 1, types["exposure"])
	require.Equal(t, 1, types["conversion"])
}

func TestClient_Close_Idempotent(t *testing.T) {
	mock := &mockHTTPClient{configResp: &SDKConfig{}}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	require.NoError(t, err)

	require.NoError(t, c.Close())
	require.NoError(t, c.Close())
}

func TestNew_InvalidConfig_ReturnsError(t *testing.T) {
	_, err := New(Config{APIKey: "", ServiceURL: "http://x"})
	require.Error(t, err)

	_, err = New(Config{APIKey: "k", ServiceURL: ""})
	require.Error(t, err)
}
