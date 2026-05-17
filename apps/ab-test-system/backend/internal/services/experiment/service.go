package experiment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

type Storage interface {
	Create(ctx context.Context, exp domain.Experiment) (domain.Experiment, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error)
	GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error)
	Update(ctx context.Context, exp domain.Experiment) (domain.Experiment, error)
	Delete(ctx context.Context, projectID, experimentID uuid.UUID) error
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(
	ctx context.Context,
	projectID uuid.UUID,
	flagID *uuid.UUID,
	name, description string,
	trafficPercent float64,
	variants []domain.Variant,
) (domain.Experiment, error) {
	exp := domain.Experiment{
		ID:             uuid.New(),
		ProjectID:      projectID,
		FlagID:         flagID,
		Name:           name,
		Description:    description,
		Status:         domain.StatusDraft,
		TrafficPercent: trafficPercent,
		Variants:       variants,
	}

	created, err := s.storage.Create(ctx, exp)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.Create: %w", err)
	}

	return created, nil
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error) {
	experiments, err := s.storage.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("experiment.Service.List: %w", err)
	}

	return experiments, nil
}

func (s *Service) GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	exp, err := s.storage.GetByID(ctx, projectID, experimentID)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.GetByID: %w", err)
	}

	return exp, nil
}

func (s *Service) Update(
	ctx context.Context,
	projectID, experimentID uuid.UUID,
	name *string,
	description *string,
	trafficPercent *float64,
) (domain.Experiment, error) {
	exp, err := s.storage.GetByID(ctx, projectID, experimentID)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.Update: %w", err)
	}

	if exp.Status != domain.StatusDraft {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.Update: %w", domain.ErrNotDraft)
	}

	if name != nil {
		exp.Name = *name
	}

	if description != nil {
		exp.Description = *description
	}

	if trafficPercent != nil {
		exp.TrafficPercent = *trafficPercent
	}

	updated, err := s.storage.Update(ctx, exp)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.Update: %w", err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, projectID, experimentID uuid.UUID) error {
	exp, err := s.storage.GetByID(ctx, projectID, experimentID)
	if err != nil {
		return fmt.Errorf("experiment.Service.Delete: %w", err)
	}

	if exp.Status != domain.StatusDraft {
		return fmt.Errorf("experiment.Service.Delete: %w", domain.ErrNotDraft)
	}

	if err := s.storage.Delete(ctx, projectID, experimentID); err != nil {
		return fmt.Errorf("experiment.Service.Delete: %w", err)
	}

	return nil
}

func (s *Service) transition(
	ctx context.Context,
	projectID, experimentID uuid.UUID,
	next domain.Status,
) (domain.Experiment, error) {
	exp, err := s.storage.GetByID(ctx, projectID, experimentID)
	if err != nil {
		return domain.Experiment{}, err
	}

	if !exp.CanTransitionTo(next) {
		return domain.Experiment{}, fmt.Errorf(
			"experiment.Service: %w: %s → %s", domain.ErrInvalidTransition, exp.Status, next,
		)
	}

	now := time.Now()
	exp.Status = next

	switch next {
	case domain.StatusRunning:
		if exp.StartedAt == nil {
			exp.StartedAt = &now
		}
	case domain.StatusCompleted:
		exp.EndedAt = &now
	}

	updated, err := s.storage.Update(ctx, exp)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("experiment.Service.transition: %w", err)
	}

	return updated, nil
}

func (s *Service) Start(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	return s.transition(ctx, projectID, experimentID, domain.StatusRunning)
}

func (s *Service) Pause(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	return s.transition(ctx, projectID, experimentID, domain.StatusPaused)
}

func (s *Service) Resume(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	return s.transition(ctx, projectID, experimentID, domain.StatusRunning)
}

func (s *Service) Complete(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	return s.transition(ctx, projectID, experimentID, domain.StatusCompleted)
}
