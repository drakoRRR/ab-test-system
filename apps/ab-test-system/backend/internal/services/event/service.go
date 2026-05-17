package event

import (
	"context"
	"fmt"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
)

type Publisher interface {
	Publish(ctx context.Context, events []domainevent.Event) error
}

type Service struct {
	publisher Publisher
}

func NewService(publisher Publisher) *Service {
	return &Service{publisher: publisher}
}

func (s *Service) Ingest(ctx context.Context, events []domainevent.Event) error {
	if err := s.publisher.Publish(ctx, events); err != nil {
		return fmt.Errorf("event.Service.Ingest: %w", err)
	}
	return nil
}
