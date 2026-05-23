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
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

var fixedResult = domainanalytics.ExperimentResult{
	ExperimentID:     fixedExperimentID,
	TotalExposures:   2000,
	TotalConversions: 700,
	Variants: []domainanalytics.VariantResult{
		{
			VariantID:      fixedControlID,
			VariantKey:     "control",
			VariantName:    "Control",
			Exposures:      1000,
			Conversions:    300,
			ConversionRate: 0.30,
			IsControl:      true,
		},
		{
			VariantID:      fixedTreatmentID,
			VariantKey:     "treatment",
			VariantName:    "Treatment",
			Exposures:      1000,
			Conversions:    400,
			ConversionRate: 0.40,
			IsControl:      false,
			Uplift:         ptr(33.33),
			PValue:         ptr(0.001),
			CILow:          ptr(0.05),
			CIHigh:         ptr(0.15),
			Significant:    ptr(true),
		},
	},
}

func TestAnalyticsHandler_GetExperimentAnalytics(t *testing.T) {
	type testCase struct {
		name       string
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.GetExperimentAnalyticsResponseObject)
	}

	tests := []testCase{
		{
			name: "200 returns analytics with correct mapping",
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetResult(mock.Anything, fixedProjectID, fixedExperimentID).
					Return(fixedResult, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetExperimentAnalyticsResponseObject) {
				r, ok := resp.(gen.GetExperimentAnalytics200JSONResponse)
				require.True(t, ok)

				assert.Equal(t, fixedExperimentID, r.ExperimentId)
				assert.Equal(t, int64(2000), r.TotalExposures)
				assert.Equal(t, int64(700), r.TotalConversions)
				require.Len(t, r.Variants, 2)

				ctrl := r.Variants[0]
				assert.Equal(t, fixedControlID, ctrl.VariantId)
				assert.Equal(t, "control", ctrl.VariantKey)
				assert.True(t, ctrl.IsControl)
				assert.InDelta(t, 0.30, ctrl.ConversionRate, 1e-9)
				assert.Nil(t, ctrl.PValue)
				assert.Nil(t, ctrl.Uplift)

				treat := r.Variants[1]
				assert.Equal(t, fixedTreatmentID, treat.VariantId)
				assert.False(t, treat.IsControl)
				assert.InDelta(t, 0.40, treat.ConversionRate, 1e-9)
				require.NotNil(t, treat.PValue)
				assert.InDelta(t, 0.001, *treat.PValue, 1e-9)
				require.NotNil(t, treat.Significant)
				assert.True(t, *treat.Significant)
			},
		},
		{
			name: "404 when experiment not found",
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetResult(mock.Anything, fixedProjectID, fixedExperimentID).
					Return(domainanalytics.ExperimentResult{}, domainexperiment.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetExperimentAnalyticsResponseObject) {
				assert.IsType(t, gen.GetExperimentAnalytics404JSONResponse{}, resp)
			},
		},
		{
			name: "service error propagates as 500",
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetResult(mock.Anything, fixedProjectID, fixedExperimentID).
					Return(domainanalytics.ExperimentResult{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetExperimentAnalytics(context.Background(), gen.GetExperimentAnalyticsRequestObject{
				ProjectId:    fixedProjectID,
				ExperimentId: fixedExperimentID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
