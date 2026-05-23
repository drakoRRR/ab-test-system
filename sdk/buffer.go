package sdk

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

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

func (b *eventBuffer) add(e event, log zerolog.Logger) {
	select {
	case b.ch <- e:
	default:
		log.Warn().Str("event_type", e.Type).Msg("sdk: event buffer full, dropping event")
	}
}

func newEvent(typ, userID, experimentID, variantID, name string, value float64) event {
	return event{
		ID:           uuid.NewString(),
		UserID:       userID,
		ExperimentID: experimentID,
		VariantID:    variantID,
		Type:         typ,
		Name:         name,
		Value:        value,
		Timestamp:    timeNow(),
	}
}

var timeNow = func() time.Time { return time.Now().UTC() }

func (cl *Client) flushLoop(ctx context.Context) {
	defer cl.wg.Done()

	ticker := time.NewTicker(cl.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]event, 0, cl.cfg.FlushBatchSize)

	for {
		select {
		case e := <-cl.buffer.ch:
			batch = append(batch, e)
			if len(batch) >= cl.cfg.FlushBatchSize {
				cl.sendBatch(ctx, batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				cl.sendBatch(ctx, batch)
				batch = batch[:0]
			}

		case <-ctx.Done():
			for len(cl.buffer.ch) > 0 {
				batch = append(batch, <-cl.buffer.ch)
			}
			if len(batch) > 0 {
				cl.sendBatch(context.Background(), batch)
			}
			return
		}
	}
}

func (cl *Client) sendBatch(ctx context.Context, batch []event) {
	if err := cl.httpCli.SendEvents(ctx, cl.cfg.APIKey, batch); err != nil {
		cl.cfg.Logger.Warn().Err(err).Int("count", len(batch)).Msg("sdk: failed to send events")
	}
}
