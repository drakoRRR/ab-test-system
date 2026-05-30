package sdk

import (
	"fmt"
	"testing"
)

var benchExp = Experiment{
	ID:             "bench-exp",
	Key:            "bench-exp",
	TrafficPercent: 100,
	Variants: []Variant{
		{ID: "v1", Key: "control", Weight: 50},
		{ID: "v2", Key: "treatment", Weight: 50},
	},
}

// BenchmarkGetVariant measures the cost of a single local evaluation.
// The SDK holds the config in a sync.RWMutex-protected cache; the call
// is a read-lock + pure MurmurHash3 computation — no network, no I/O.
//
// Run: cd sdk && go test -bench=BenchmarkGetVariant -benchmem -count=5
func BenchmarkGetVariant(b *testing.B) {
	mock := &mockHTTPClient{
		configResp: &SDKConfig{Experiments: []Experiment{benchExp}},
	}
	c, err := newWithHTTPClient(Config{APIKey: "k", ServiceURL: "http://x"}, mock)
	if err != nil {
		b.Fatal(err)
	}
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.GetVariant("bench-exp", fmt.Sprintf("user-%d", i))
			i++
		}
	})
}

// BenchmarkAssignBucket isolates the raw hash computation without
// any locking or struct overhead.
func BenchmarkAssignBucket(b *testing.B) {
	b.ResetTimer()
	for i := range b.N {
		assignBucket(fmt.Sprintf("user-%d", i), "bench-key")
	}
}
