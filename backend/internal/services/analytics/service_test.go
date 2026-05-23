package analytics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainanalytics "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/analytics"
	domainexperiment "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

func TestService_GetResult(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		check     func(t *testing.T, result domainanalytics.ExperimentResult)
	}

	tests := []testCase{
		{
			name: "happy path — two variants, treatment receives stats",
			setup: func(ms *mockedService) {
				ms.experiments.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(baseExperiment, nil)
				ms.metrics.EXPECT().GetMetrics(mock.Anything, experimentID).Return([]domainanalytics.VariantMetrics{
					{VariantID: controlID, Exposures: 1000, Conversions: 300},
					{VariantID: treatmentID, Exposures: 1000, Conversions: 400},
				}, nil)
			},
			assertErr: assert.NoError,
			check: func(t *testing.T, r domainanalytics.ExperimentResult) {
				assert.Equal(t, experimentID, r.ExperimentID)
				assert.Equal(t, int64(2000), r.TotalExposures)
				assert.Equal(t, int64(700), r.TotalConversions)
				require.Len(t, r.Variants, 2)

				ctrl := r.Variants[0]
				assert.True(t, ctrl.IsControl)
				assert.InDelta(t, 0.30, ctrl.ConversionRate, 1e-9)
				assert.Nil(t, ctrl.PValue)

				treat := r.Variants[1]
				assert.False(t, treat.IsControl)
				assert.InDelta(t, 0.40, treat.ConversionRate, 1e-9)
				assert.NotNil(t, treat.PValue)
				assert.NotNil(t, treat.Uplift)
				assert.Greater(t, *treat.Uplift, 0.0)
			},
		},
		{
			name: "experiment not found",
			setup: func(ms *mockedService) {
				ms.experiments.EXPECT().GetByID(mock.Anything, projectID, experimentID).
					Return(domainexperiment.Experiment{}, domainexperiment.ErrNotFound)
			},
			assertErr: errIs(domainexperiment.ErrNotFound),
		},
		{
			name: "metrics storage error propagates",
			setup: func(ms *mockedService) {
				ms.experiments.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(baseExperiment, nil)
				ms.metrics.EXPECT().GetMetrics(mock.Anything, experimentID).Return(nil, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
		{
			name: "no events yet — treatment has nil stats",
			setup: func(ms *mockedService) {
				ms.experiments.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(baseExperiment, nil)
				ms.metrics.EXPECT().GetMetrics(mock.Anything, experimentID).Return(nil, nil)
			},
			assertErr: assert.NoError,
			check: func(t *testing.T, r domainanalytics.ExperimentResult) {
				assert.Equal(t, int64(0), r.TotalExposures)
				assert.Equal(t, int64(0), r.TotalConversions)
				require.Len(t, r.Variants, 2)
				assert.Nil(t, r.Variants[1].PValue)
			},
		},
		{
			name: "single variant (control only) — no stats computed",
			setup: func(ms *mockedService) {
				controlOnly := baseExperiment
				controlOnly.Variants = baseExperiment.Variants[:1]
				ms.experiments.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(controlOnly, nil)
				ms.metrics.EXPECT().GetMetrics(mock.Anything, experimentID).Return([]domainanalytics.VariantMetrics{
					{VariantID: controlID, Exposures: 500, Conversions: 100},
				}, nil)
			},
			assertErr: assert.NoError,
			check: func(t *testing.T, r domainanalytics.ExperimentResult) {
				require.Len(t, r.Variants, 1)
				assert.True(t, r.Variants[0].IsControl)
				assert.Nil(t, r.Variants[0].PValue)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setup(ms)

			result, err := ms.GetResult(context.Background(), projectID, experimentID)

			tc.assertErr(t, err)
			if err == nil && tc.check != nil {
				tc.check(t, result)
			}
		})
	}
}
