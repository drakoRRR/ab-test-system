package analytics

import (
	"context"

	"github.com/google/uuid"

	domainanalytics "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/analytics"
)

type Service interface {
	GetResult(ctx context.Context, projectID, experimentID uuid.UUID) (domainanalytics.ExperimentResult, error)
}

type AnalyticsHandler struct {
	service Service
}

func NewHandler(service Service) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}
