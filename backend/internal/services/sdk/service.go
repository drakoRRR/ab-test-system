package sdk

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
)

type FlagLister interface {
	List(ctx context.Context, projectID uuid.UUID) ([]domainflag.Flag, error)
}

type ExperimentLister interface {
	List(ctx context.Context, projectID uuid.UUID) ([]domainexp.Experiment, error)
}

type Service struct {
	flags       FlagLister
	experiments ExperimentLister
}

func NewService(flags FlagLister, experiments ExperimentLister) *Service {
	return &Service{flags: flags, experiments: experiments}
}

func (s *Service) GetConfig(ctx context.Context, projectID uuid.UUID) (domainsdk.Config, error) {
	flags, err := s.flags.List(ctx, projectID)
	if err != nil {
		return domainsdk.Config{}, fmt.Errorf("sdk.Service.GetConfig: flags: %w", err)
	}

	all, err := s.experiments.List(ctx, projectID)
	if err != nil {
		return domainsdk.Config{}, fmt.Errorf("sdk.Service.GetConfig: experiments: %w", err)
	}

	running := make([]domainexp.Experiment, 0, len(all))
	for _, e := range all {
		if e.Status == domainexp.StatusRunning {
			running = append(running, e)
		}
	}

	return domainsdk.Config{
		ProjectID:   projectID,
		Flags:       flags,
		Experiments: running,
	}, nil
}
