package experiment

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

type Service interface {
	Create(
		ctx context.Context,
		projectID uuid.UUID,
		flagID *uuid.UUID,
		name, description string,
		trafficPercent float64,
		variants []domain.Variant,
	) (domain.Experiment, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error)
	GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
	Update(
		ctx context.Context,
		projectID, experimentID uuid.UUID,
		name *string,
		description *string,
		trafficPercent *float64,
	) (domain.Experiment, error)
	Delete(ctx context.Context, projectID, experimentID uuid.UUID) error
	Start(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
	Pause(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
	Resume(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
	Complete(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
}

type ExperimentHandler struct {
	service Service
}

func NewHandler(service Service) *ExperimentHandler {
	return &ExperimentHandler{service: service}
}
