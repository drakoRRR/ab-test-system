package sdk

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockHTTPClient is a test double that records sent events.
// Also used in sdk_test.go.
type mockHTTPClient struct {
	mu         sync.Mutex
	sentEvents []event
	configResp *SDKConfig
}

func (m *mockHTTPClient) FetchConfig(_ context.Context, _ string) (*SDKConfig, error) {
	if m.configResp != nil {
		return m.configResp, nil
	}
	return &SDKConfig{}, nil
}

func (m *mockHTTPClient) SendEvents(_ context.Context, _ string, batch []event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentEvents = append(m.sentEvents, batch...)
	return nil
}

func (m *mockHTTPClient) events() []event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]event, len(m.sentEvents))
	copy(out, m.sentEvents)
	return out
}

func TestBuffer_FlushOnSizeThreshold(t *testing.T) {
	cfg := Config{
		APIKey:         "test",
		ServiceURL:     "http://x",
		FlushInterval:  10 * time.Second,
		FlushBatchSize: 3,
		MaxBufferSize:  100,
	}.withDefaults()

	mock := &mockHTTPClient{}
	buf := newEventBuffer(cfg)

	cl := &Client{cfg: cfg, buffer: buf, httpCli: mock}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl.wg.Add(1)
	go cl.flushLoop(ctx)

	for i := 0; i < 3; i++ {
		buf.add(event{ID: "e", Type: "exposure", UserID: "u"}, cfg.Logger)
	}

	require.Eventually(t, func() bool {
		return len(mock.events()) >= 3
	}, 2*time.Second, 50*time.Millisecond)
}

func TestBuffer_FlushOnTimeThreshold(t *testing.T) {
	cfg := Config{
		APIKey:         "test",
		ServiceURL:     "http://x",
		FlushInterval:  100 * time.Millisecond,
		FlushBatchSize: 100,
		MaxBufferSize:  100,
	}.withDefaults()

	mock := &mockHTTPClient{}
	buf := newEventBuffer(cfg)

	cl := &Client{cfg: cfg, buffer: buf, httpCli: mock}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl.wg.Add(1)
	go cl.flushLoop(ctx)

	buf.add(event{ID: "e", Type: "exposure", UserID: "u"}, cfg.Logger)

	require.Eventually(t, func() bool {
		return len(mock.events()) >= 1
	}, 2*time.Second, 50*time.Millisecond)
}

func TestBuffer_DropsWhenFull(t *testing.T) {
	cfg := Config{
		APIKey:        "test",
		ServiceURL:    "http://x",
		MaxBufferSize: 2,
		FlushInterval: 10 * time.Second,
	}.withDefaults()

	buf := newEventBuffer(cfg)
	logger := cfg.Logger

	buf.add(event{ID: "1"}, logger)
	buf.add(event{ID: "2"}, logger)
	buf.add(event{ID: "3"}, logger) // dropped silently

	require.Equal(t, 2, len(buf.ch))
}
