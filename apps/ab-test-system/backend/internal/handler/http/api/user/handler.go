package user

import (
	"context"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type Service interface {
	CreateOrUpdate(ctx context.Context, firebaseUID, email, name string, photoURL *string) (domain.User, error)
	GetCurrentUser(ctx context.Context, firebaseUID string) (domain.User, error)
}

type UserHandler struct {
	service Service
}

func NewHandler(service Service) *UserHandler {
	return &UserHandler{service: service}
}
