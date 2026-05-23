package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, cfg ProducerConfig) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, events []domainevent.Event) error {
	msgs := make([]kafka.Message, len(events))
	for i, e := range events {
		val, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("kafka.Producer.Publish: marshal: %w", err)
		}
		msgs[i] = kafka.Message{
			Key:   []byte(e.ID.String()),
			Value: val,
		}
	}

	if err := p.writer.WriteMessages(ctx, msgs...); err != nil {
		return fmt.Errorf("kafka.Producer.Publish: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
