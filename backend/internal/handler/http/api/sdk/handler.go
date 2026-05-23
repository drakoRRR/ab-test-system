package sdk

import (
	"context"

	"github.com/google/uuid"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
)

type ConfigService interface {
	GetConfig(ctx context.Context, projectID uuid.UUID) (domainsdk.Config, error)
}

type EventService interface {
	Ingest(ctx context.Context, events []domainevent.Event) error
}

type Handler struct {
	config ConfigService
	events EventService
}

func NewHandler(config ConfigService, events EventService) *Handler {
	return &Handler{config: config, events: events}
}
