package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
)

type EventStore interface {
	BulkInsert(ctx context.Context, events []domainevent.Event) error
}

type Consumer struct {
	reader *kafka.Reader
	store  EventStore
	log    zerolog.Logger

	batchSize    int
	flushTimeout time.Duration
}

func NewConsumer(brokers []string, cfg ConsumerConfig, store EventStore, log zerolog.Logger) *Consumer {
	cfg = cfg.withDefaults()
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
		store:        store,
		log:          log,
		batchSize:    cfg.BatchSize,
		flushTimeout: cfg.FlushTimeout,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error().Err(err).Msg("consumer: reader close failed")
		}
	}()

	batch := make([]domainevent.Event, 0, c.batchSize)

	for {
		fetchCtx, cancel := context.WithTimeout(ctx, c.flushTimeout)
		msg, err := c.reader.FetchMessage(fetchCtx)
		cancel()

		if err != nil {
			c.flushBatch(ctx, batch)
			batch = batch[:0]
			if ctx.Err() != nil {
				return nil
			}
			continue // flushTimeout elapsed with no message — flush and wait again
		}

		var e domainevent.Event
		if jsonErr := json.Unmarshal(msg.Value, &e); jsonErr != nil {
			c.log.Warn().Err(jsonErr).Msg("consumer: unmarshal failed, skipping")
		} else {
			batch = append(batch, e)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("consumer: commit: %w", err)
		}

		if len(batch) >= c.batchSize {
			c.flushBatch(ctx, batch)
			batch = batch[:0]
		}
	}
}

func (c *Consumer) flushBatch(ctx context.Context, batch []domainevent.Event) {
	if len(batch) == 0 {
		return
	}
	if err := c.store.BulkInsert(ctx, batch); err != nil {
		c.log.Error().Err(err).Int("count", len(batch)).Msg("consumer: bulk insert failed")
	} else {
		c.log.Debug().Int("count", len(batch)).Msg("consumer: flushed batch")
	}
}
