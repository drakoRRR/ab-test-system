package stats_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/stats"
)

func TestTwoProportions_ReturnsNil(t *testing.T) {
	tests := []struct {
		name                 string
		n1, conv1, n2, conv2 int64
	}{
		{"zero control exposures", 0, 0, 100, 30},
		{"zero treatment exposures", 100, 30, 0, 0},
		{"zero conversions in both groups (pooled p=0)", 100, 0, 100, 0},
		{"full conversion in both groups (pooled p=1)", 100, 100, 100, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, stats.TwoProportions(tt.n1, tt.conv1, tt.n2, tt.conv2))
		})
	}
}

func TestTwoProportions(t *testing.T) {
	tests := []struct {
		name                 string
		n1, conv1, n2, conv2 int64
		check                func(t *testing.T, r *stats.TwoProportionResult)
	}{
		{
			name: "identical rates — z=0, p=1, not significant, zero uplift",
			n1:   1000, conv1: 300, n2: 1000, conv2: 300,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				assert.InDelta(t, 0.0, r.ZStat, 1e-9)
				assert.InDelta(t, 1.0, r.PValue, 1e-9)
				assert.False(t, r.Significant)
				assert.InDelta(t, 0.0, r.Uplift, 1e-9)
			},
		},
		{
			name: "significant positive difference — 30% vs 40% at n=5000",
			n1:   5000, conv1: 1500, n2: 5000, conv2: 2000,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				assert.True(t, r.Significant)
				assert.Less(t, r.PValue, 0.05)
				assert.Greater(t, r.Uplift, 0.0)
				assert.Greater(t, r.CILow, -1.0) // sanity bounds
				assert.Less(t, r.CIHigh, 1.0)
			},
		},
		{
			name: "negative uplift — treatment worse than control",
			n1:   5000, conv1: 2000, n2: 5000, conv2: 1500,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				assert.Less(t, r.Uplift, 0.0)
				assert.Less(t, r.ZStat, 0.0)
			},
		},
		{
			name: "non-significant result — CI straddles zero",
			n1:   50, conv1: 15, n2: 50, conv2: 18,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				if !r.Significant {
					assert.True(t, r.CILow < 0 && r.CIHigh > 0, "CI must straddle zero for non-significant result")
				}
			},
		},
		{
			name: "known uplift — 20% vs 25% → +25%",
			n1:   10000, conv1: 2000, n2: 10000, conv2: 2500,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				assert.InDelta(t, 25.0, r.Uplift, 0.1)
			},
		},
		{
			name: "p-value is in [0,1] and not NaN",
			n1:   1000, conv1: 300, n2: 1000, conv2: 400,
			check: func(t *testing.T, r *stats.TwoProportionResult) {
				assert.False(t, math.IsNaN(r.PValue))
				assert.GreaterOrEqual(t, r.PValue, 0.0)
				assert.LessOrEqual(t, r.PValue, 1.0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := stats.TwoProportions(tt.n1, tt.conv1, tt.n2, tt.conv2)
			require.NotNil(t, r)
			tt.check(t, r)
		})
	}
}
