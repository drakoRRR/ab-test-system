package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domainanalytics "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/analytics"
	domainexperiment "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/stats"
)

type MetricsStorage interface {
	GetMetrics(ctx context.Context, experimentID uuid.UUID) ([]domainanalytics.VariantMetrics, error)
}

type ExperimentStorage interface {
	GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domainexperiment.Experiment, error)
}

type Service struct {
	metrics     MetricsStorage
	experiments ExperimentStorage
}

func NewService(metrics MetricsStorage, experiments ExperimentStorage) *Service {
	return &Service{metrics: metrics, experiments: experiments}
}

func (s *Service) GetResult(
	ctx context.Context,
	projectID, experimentID uuid.UUID,
) (domainanalytics.ExperimentResult, error) {
	exp, err := s.experiments.GetByID(ctx, projectID, experimentID)
	if err != nil {
		return domainanalytics.ExperimentResult{}, fmt.Errorf("analytics.Service.GetResult: %w", err)
	}

	rawMetrics, err := s.metrics.GetMetrics(ctx, experimentID)
	if err != nil {
		return domainanalytics.ExperimentResult{}, fmt.Errorf("analytics.Service.GetResult: %w", err)
	}

	metricsByVariant := make(map[uuid.UUID]domainanalytics.VariantMetrics, len(rawMetrics))
	for _, m := range rawMetrics {
		metricsByVariant[m.VariantID] = m
	}

	var controlMetrics domainanalytics.VariantMetrics
	if len(exp.Variants) > 0 {
		controlMetrics = metricsByVariant[exp.Variants[0].ID]
	}

	result := domainanalytics.ExperimentResult{
		ExperimentID: experimentID,
	}

	for i, v := range exp.Variants {
		m := metricsByVariant[v.ID]

		rate := 0.0
		if m.Exposures > 0 {
			rate = float64(m.Conversions) / float64(m.Exposures)
		}

		vr := domainanalytics.VariantResult{
			VariantID:      v.ID,
			VariantKey:     v.Key,
			VariantName:    v.Name,
			Exposures:      m.Exposures,
			Conversions:    m.Conversions,
			ConversionRate: rate,
			IsControl:      i == 0,
		}

		if i > 0 {
			if r := stats.TwoProportions(
				controlMetrics.Exposures, controlMetrics.Conversions,
				m.Exposures, m.Conversions,
			); r != nil {
				vr.Uplift = &r.Uplift
				vr.PValue = &r.PValue
				vr.CILow = &r.CILow
				vr.CIHigh = &r.CIHigh
				vr.Significant = &r.Significant
			}
		}

		result.TotalExposures += m.Exposures
		result.TotalConversions += m.Conversions
		result.Variants = append(result.Variants, vr)
	}

	return result, nil
}
