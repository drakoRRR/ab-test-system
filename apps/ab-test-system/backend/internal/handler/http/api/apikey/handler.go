package apikey

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
)

type Service interface {
	Create(ctx context.Context, projectID uuid.UUID, name string) (domain.Key, string, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Key, error)
	Revoke(ctx context.Context, projectID, keyID uuid.UUID) error
}

type APIKeyHandler struct {
	service Service
}

func NewHandler(service Service) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}
