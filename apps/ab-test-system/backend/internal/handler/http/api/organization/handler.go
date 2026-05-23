package organization

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/organization"
	domainuser "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type Service interface {
	Create(ctx context.Context, name string, userID uuid.UUID) (domain.Organization, error)
}

type UserService interface {
	GetCurrentUser(ctx context.Context, firebaseUID string) (domainuser.User, error)
}

type Handler struct {
	service     Service
	userService UserService
}

func NewHandler(service Service, userService UserService) *Handler {
	return &Handler{service: service, userService: userService}
}
