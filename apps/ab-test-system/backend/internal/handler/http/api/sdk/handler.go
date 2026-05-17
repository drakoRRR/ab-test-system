package sdk

import (
	"context"

	"github.com/google/uuid"

	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
)

type Service interface {
	GetConfig(ctx context.Context, projectID uuid.UUID) (domainsdk.Config, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
