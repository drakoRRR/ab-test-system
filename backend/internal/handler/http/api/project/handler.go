package project

import (
	"context"

	"github.com/google/uuid"

	domainproject "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	domainuser "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type UserService interface {
	GetCurrentUser(ctx context.Context, firebaseUID string) (domainuser.User, error)
}

type ProjectService interface {
	Create(ctx context.Context, orgID uuid.UUID, name, description string) (domainproject.Project, error)
	List(ctx context.Context, orgID uuid.UUID) ([]domainproject.Project, error)
	GetByID(ctx context.Context, orgID, projectID uuid.UUID) (domainproject.Project, error)
	Update(ctx context.Context, p domainproject.UpdateParams) (domainproject.Project, error)
	Delete(ctx context.Context, orgID, projectID uuid.UUID) error
}

type ProjectHandler struct {
	users    UserService
	projects ProjectService
}

func NewHandler(users UserService, projects ProjectService) *ProjectHandler {
	return &ProjectHandler{users: users, projects: projects}
}
