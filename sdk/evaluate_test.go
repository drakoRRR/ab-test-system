package sdk

import (
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
