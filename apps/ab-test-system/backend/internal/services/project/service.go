package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
)

type Storage interface {
	Create(ctx context.Context, p domain.Project) (domain.Project, error)
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error)
	GetByID(ctx context.Context, id, orgID uuid.UUID) (domain.Project, error)
	Update(ctx context.Context, p domain.Project) (domain.Project, error)
	Delete(ctx context.Context, id, orgID uuid.UUID) error
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(ctx context.Context, orgID uuid.UUID, name, description string) (domain.Project, error) {
	p := domain.Project{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        name,
		Description: description,
	}

	created, err := s.storage.Create(ctx, p)
	if err != nil {
		return domain.Project{}, fmt.Errorf("project.Service.Create: %w", err)
	}

	return created, nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error) {
	projects, err := s.storage.List(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("project.Service.List: %w", err)
	}

	return projects, nil
}

func (s *Service) GetByID(ctx context.Context, orgID, projectID uuid.UUID) (domain.Project, error) {
	p, err := s.storage.GetByID(ctx, projectID, orgID)
	if err != nil {
		return domain.Project{}, fmt.Errorf("project.Service.GetByID: %w", err)
	}

	return p, nil
}

func (s *Service) Update(
	ctx context.Context,
	orgID, projectID uuid.UUID,
	name, description *string,
) (domain.Project, error) {
	p, err := s.storage.GetByID(ctx, projectID, orgID)
	if err != nil {
		return domain.Project{}, fmt.Errorf("project.Service.Update: %w", err)
	}

	if name != nil {
		p.Name = *name
	}

	if description != nil {
		p.Description = *description
	}

	updated, err := s.storage.Update(ctx, p)
	if err != nil {
		return domain.Project{}, fmt.Errorf("project.Service.Update: %w", err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, orgID, projectID uuid.UUID) error {
	if err := s.storage.Delete(ctx, projectID, orgID); err != nil {
		return fmt.Errorf("project.Service.Delete: %w", err)
	}

	return nil
}
