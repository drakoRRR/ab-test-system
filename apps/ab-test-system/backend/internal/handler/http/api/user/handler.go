package user

import (
	"context"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type Service interface {
	CreateOrUpdate(ctx context.Context, firebaseUID, email, name string, photoURL *string) (domain.User, error)
	GetCurrentUser(ctx context.Context, firebaseUID string) (domain.User, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
