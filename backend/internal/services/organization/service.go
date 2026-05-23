package organization

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/organization"
)

type Storage interface {
	// Create inserts a new organization and sets org_id on the user in one transaction.
	Create(ctx context.Context, name string, userID uuid.UUID) (domain.Organization, error)
}

type UserStorage interface {
	HasOrg(ctx context.Context, userID uuid.UUID) (bool, error)
}

type Service struct {
	storage     Storage
	userStorage UserStorage
}

func NewService(storage Storage, userStorage UserStorage) *Service {
	return &Service{storage: storage, userStorage: userStorage}
}

func (s *Service) Create(ctx context.Context, name string, userID uuid.UUID) (domain.Organization, error) {
	has, err := s.userStorage.HasOrg(ctx, userID)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("organization.Service.Create: %w", err)
	}
	if has {
		return domain.Organization{}, fmt.Errorf("organization.Service.Create: %w", domain.ErrAlreadyHasOrg)
	}

	org, err := s.storage.Create(ctx, name, userID)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("organization.Service.Create: %w", err)
	}

	return org, nil
}
