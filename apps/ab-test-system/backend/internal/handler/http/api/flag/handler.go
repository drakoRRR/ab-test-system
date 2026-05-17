package flag

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

type Service interface {
	Create(ctx context.Context, projectID uuid.UUID, key, name string) (domain.Flag, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Flag, error)
	GetByKey(ctx context.Context, projectID uuid.UUID, key string) (domain.Flag, error)
	Update(
		ctx context.Context,
		projectID uuid.UUID,
		key string,
		name *string,
		enabled *bool,
		rules *[]domain.Rule,
	) (domain.Flag, error)
	Delete(ctx context.Context, projectID uuid.UUID, key string) error
}

type FlagHandler struct {
	service Service
}

func NewHandler(service Service) *FlagHandler {
	return &FlagHandler{service: service}
}
