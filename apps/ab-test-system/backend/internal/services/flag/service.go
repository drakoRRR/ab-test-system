package flag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

type Storage interface {
	Create(ctx context.Context, f domain.Flag) (domain.Flag, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Flag, error)
	GetByKey(ctx context.Context, projectID uuid.UUID, key string) (domain.Flag, error)
	Update(ctx context.Context, f domain.Flag) (domain.Flag, error)
	Delete(ctx context.Context, projectID uuid.UUID, key string) error
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(ctx context.Context, projectID uuid.UUID, key, name string) (domain.Flag, error) {
	f := domain.Flag{
		ID:        uuid.New(),
		ProjectID: projectID,
		Key:       key,
		Name:      name,
		Enabled:   false,
		Rules:     []domain.Rule{},
	}

	created, err := s.storage.Create(ctx, f)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("flag.Service.Create: %w", err)
	}

	return created, nil
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID) ([]domain.Flag, error) {
	flags, err := s.storage.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("flag.Service.List: %w", err)
	}

	return flags, nil
}

func (s *Service) GetByKey(ctx context.Context, projectID uuid.UUID, key string) (domain.Flag, error) {
	f, err := s.storage.GetByKey(ctx, projectID, key)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("flag.Service.GetByKey: %w", err)
	}

	return f, nil
}

func (s *Service) Update(
	ctx context.Context,
	projectID uuid.UUID,
	key string,
	name *string,
	enabled *bool,
	rules *[]domain.Rule,
) (domain.Flag, error) {
	f, err := s.storage.GetByKey(ctx, projectID, key)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("flag.Service.Update: %w", err)
	}

	if name != nil {
		f.Name = *name
	}

	if enabled != nil {
		f.Enabled = *enabled
	}

	if rules != nil {
		f.Rules = *rules
	}

	updated, err := s.storage.Update(ctx, f)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("flag.Service.Update: %w", err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, projectID uuid.UUID, key string) error {
	if err := s.storage.Delete(ctx, projectID, key); err != nil {
		return fmt.Errorf("flag.Service.Delete: %w", err)
	}

	return nil
}
