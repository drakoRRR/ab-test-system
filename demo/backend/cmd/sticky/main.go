// Scenario D: Sticky bucketing and distribution uniformity.
//
// Runs two checks against the live demo server:
//
//   a) Consistency — 1 000 users × 5 GET /visit calls each.
//      Every call for the same userID must return the same variant (100% expected).
//
//   c) Distribution uniformity — counts control vs treatment across those 1 000 users
//      and runs a two-proportion z-test on H₀: p = 0.5 (equal split).
//      Expected: p-value > 0.05 (cannot reject equal distribution).
//
// Usage:
//
//	go run ./demo/cmd/sticky
//
// Environment variables:
//
//	DEMO_URL        (default http://localhost:8081)
//	EXPERIMENT_KEY  (default checkout-btn)
//	N_USERS         (default 1000)
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	demoURL := envOr("DEMO_URL", "http://localhost:8081")
	expKey := envOr("EXPERIMENT_KEY", "checkout-btn")
	nUsers := envIntOr("N_USERS", 1000)

	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Printf("Sticky bucketing test\n")
	fmt.Printf("  demo URL:        %s\n", demoURL)
	fmt.Printf("  experiment key:  %s\n", expKey)
	fmt.Printf("  users:           %d\n\n", nUsers)

	counts := map[string]int{"control": 0, "treatment": 0}
	inconsistent := 0

	for i := range nUsers {
		uid := fmt.Sprintf("sticky-user-%06d", i)
		var first string

		for call := range 5 {
			variant, err := getVariant(client, demoURL, uid, expKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  user %s call %d: %v\n", uid, call, err)
				continue
			}
			if call == 0 {
				first = variant
				counts[variant]++
			} else if variant != first {
				inconsistent++
				fmt.Fprintf(os.Stderr,
					"  INCONSISTENT user=%s call=%d expected=%q got=%q\n",
					uid, call, first, variant)
			}
		}
	}

	total := counts["control"] + counts["treatment"]

	fmt.Println("─── a) Consistency ────────────────────────────────────────────")
	fmt.Printf("  users tested:   %d × 5 calls\n", nUsers)
	fmt.Printf("  inconsistencies: %d\n", inconsistent)
	if inconsistent == 0 {
		fmt.Println("  result:          PASS — 100% sticky bucketing")
	} else {
		fmt.Printf("  result:          FAIL — %d assignment(s) changed between calls\n", inconsistent)
	}

	fmt.Println("\n─── c) Distribution uniformity ────────────────────────────────")
	fmt.Printf("  control:   %d (%.1f%%)\n", counts["control"], 100*float64(counts["control"])/float64(total))
	fmt.Printf("  treatment: %d (%.1f%%)\n", counts["treatment"], 100*float64(counts["treatment"])/float64(total))

	pVal := distributionPValue(counts["control"], total)
	fmt.Printf("  z-test (H₀: p=0.5): p-value = %.4f\n", pVal)
	if pVal > 0.05 {
		fmt.Println("  result:          PASS — cannot reject equal split (p > 0.05)")
	} else {
		fmt.Printf("  result:          FAIL — split is significantly unequal (p = %.4f)\n", pVal)
	}

	if inconsistent > 0 {
		os.Exit(1)
	}
}

// getVariant calls GET /visit and returns the assigned variant key.
func getVariant(client *http.Client, baseURL, userID, expKey string) (string, error) {
	url := fmt.Sprintf("%s/visit?user_id=%s&experiment_key=%s", baseURL, userID, expKey)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var body struct {
		Variant string `json:"variant"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Variant, nil
}

// distributionPValue runs a two-tailed z-test on H₀: p = 0.5.
// Returns the p-value; values > 0.05 mean we cannot reject the null hypothesis.
func distributionPValue(successes, total int) float64 {
	if total == 0 {
		return 1
	}
	pHat := float64(successes) / float64(total)
	se := math.Sqrt(0.5 * 0.5 / float64(total))
	z := math.Abs(pHat-0.5) / se
	return 2 * (1 - normalCDF(z))
}

func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
