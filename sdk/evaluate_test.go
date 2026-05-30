package sdk

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- assignBucket ---

func TestAssignBucket_Deterministic(t *testing.T) {
	a := assignBucket("user-123", "checkout-btn")
	b := assignBucket("user-123", "checkout-btn")
	require.Equal(t, a, b)
}

func TestAssignBucket_DifferentKey_DifferentBucket(t *testing.T) {
	a := assignBucket("user-123", "checkout-btn")
	b := assignBucket("user-123", "banner-color")
	require.NotEqual(t, a, b, "same user, different key must produce different buckets")
}

func TestAssignBucket_Range(t *testing.T) {
	for _, userID := range []string{"alice", "bob", "carol", "dave", "eve"} {
		bucket := assignBucket(userID, "some-key")
		require.Less(t, bucket, uint32(10000), "bucket must be in [0, 9999]")
	}
}

// TestAssignBucket_KnownValues pins the exact murmur3 outputs for the inputs used
// by the backend. Any change to assignBucket that breaks these values would silently
// split users into different buckets than the server — a critical consistency bug.
func TestAssignBucket_KnownValues(t *testing.T) {
	require.Equal(t, uint32(8833), assignBucket("user-123", "checkout-btn"))
	require.Equal(t, uint32(8746), assignBucket("user-456", "checkout-btn"))
	require.Equal(t, uint32(3349), assignBucket("user-123", "banner-color"))
}

// --- evaluateFlag ---

func TestEvaluateFlag_Disabled(t *testing.T) {
	flag := Flag{Key: "f", Enabled: false, Rules: nil}
	require.False(t, evaluateFlag(flag, "any-user"))
}

func TestEvaluateFlag_EnabledNoRules(t *testing.T) {
	flag := Flag{Key: "f", Enabled: true, Rules: nil}
	require.True(t, evaluateFlag(flag, "any-user"))
}

func TestEvaluateFlag_PercentageRule_100Percent(t *testing.T) {
	flag := Flag{
		Key:     "f",
		Enabled: true,
		Rules:   []FlagRule{{Type: "percentage", Value: 1.0}},
	}
	for _, uid := range []string{"a", "b", "c", "d", "e"} {
		require.True(t, evaluateFlag(flag, uid), "user %s should be included", uid)
	}
}

func TestEvaluateFlag_PercentageRule_0Percent(t *testing.T) {
	flag := Flag{
		Key:     "f",
		Enabled: true,
		Rules:   []FlagRule{{Type: "percentage", Value: 0.0}},
	}
	for _, uid := range []string{"a", "b", "c", "d", "e"} {
		require.False(t, evaluateFlag(flag, uid), "user %s should be excluded", uid)
	}
}

// --- evaluateExperiment ---

var runningExp = Experiment{
	Key:            "btn-color",
	TrafficPercent: 100,
	Variants: []Variant{
		{ID: "v1", Key: "control", Weight: 50},
		{ID: "v2", Key: "treatment", Weight: 50},
	},
}

func TestEvaluateExperiment_ReturnsVariantForAllTraffic(t *testing.T) {
	for _, uid := range []string{"alice", "bob", "carol", "dave", "eve"} {
		variant := evaluateExperiment(runningExp, uid)
		require.NotEmpty(t, variant, "user %s should be assigned a variant", uid)
		require.True(t, variant == "control" || variant == "treatment")
	}
}

func TestEvaluateExperiment_ZeroTraffic_NoVariant(t *testing.T) {
	exp := Experiment{
		Key:            "btn-color",
		TrafficPercent: 0,
		Variants:       runningExp.Variants,
	}
	for _, uid := range []string{"alice", "bob", "carol"} {
		require.Empty(t, evaluateExperiment(exp, uid))
	}
}

func TestEvaluateExperiment_Deterministic(t *testing.T) {
	v1 := evaluateExperiment(runningExp, "user-stable")
	v2 := evaluateExperiment(runningExp, "user-stable")
	require.Equal(t, v1, v2, "same user must always get the same variant")
}

func TestEvaluateExperiment_VariantDistribution(t *testing.T) {
	counts := map[string]int{}
	for i := 0; i < 1000; i++ {
		uid := "user-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		v := evaluateExperiment(runningExp, uid)
		if v != "" {
			counts[v]++
		}
	}
	total := counts["control"] + counts["treatment"]
	require.Greater(t, counts["control"], total*30/100)
	require.Greater(t, counts["treatment"], total*30/100)
}

// --- Scenario D.b: determinism across repeated calls (1000 users × 5 calls) ---

// TestEvaluateExperiment_StickyAcrossRepeatedCalls verifies that the same
// (userID, experimentKey) pair always resolves to the same variant.
// This is the algorithmic foundation of sticky bucketing.
func TestEvaluateExperiment_StickyAcrossRepeatedCalls(t *testing.T) {
	exp := Experiment{
		Key:            "sticky-test",
		TrafficPercent: 100,
		Variants: []Variant{
			{ID: "v1", Key: "control", Weight: 50},
			{ID: "v2", Key: "treatment", Weight: 50},
		},
	}

	for i := range 1000 {
		uid := fmt.Sprintf("sticky-user-%d", i)
		first := evaluateExperiment(exp, uid)
		require.NotEmpty(t, first, "user %s should be assigned a variant", uid)
		for call := 1; call < 5; call++ {
			got := evaluateExperiment(exp, uid)
			require.Equal(t, first, got,
				"user %s call %d: expected %q, got %q", uid, call, first, got)
		}
	}
}

// --- Scenario D.d: bucket uniformity chi-square test (100k users → 10k buckets) ---

// TestAssignBucket_ChiSquareUniformity generates 100 000 synthetic user IDs and
// verifies that the MurmurHash3 output is uniformly distributed across all 10 000
// buckets using a chi-square goodness-of-fit test (H₀: uniform distribution).
//
// With df = 9999 the expected chi² ≈ 9999; 3σ bounds are [9576, 10422].
// A value outside this range would indicate a systematic bias in the hash function.
func TestAssignBucket_ChiSquareUniformity(t *testing.T) {
	const (
		nUsers   = 100_000
		nBuckets = 10_000
	)

	counts := make([]int, nBuckets)
	for i := range nUsers {
		uid := fmt.Sprintf("chi-user-%d", i)
		counts[assignBucket(uid, "chi-square-test")]++
	}

	expected := float64(nUsers) / float64(nBuckets) // 10.0 per bucket
	chi2 := 0.0
	for _, c := range counts {
		diff := float64(c) - expected
		chi2 += diff * diff / expected
	}

	df := float64(nBuckets - 1)
	sigma3 := 3 * math.Sqrt(2*df)
	require.InDelta(t, df, chi2, sigma3,
		"chi² = %.1f is outside 3σ range [%.0f, %.0f] for df=%.0f — bucket distribution is not uniform",
		chi2, df-sigma3, df+sigma3, df)
}
