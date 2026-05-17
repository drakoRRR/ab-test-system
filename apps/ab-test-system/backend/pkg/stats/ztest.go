package stats

import "math"

// TwoProportionResult holds the output of a two-proportion z-test.
type TwoProportionResult struct {
	ZStat       float64
	PValue      float64
	CILow       float64 // 95% confidence interval lower bound for (p2 - p1)
	CIHigh      float64 // 95% confidence interval upper bound for (p2 - p1)
	Uplift      float64 // (p2 - p1) / p1 * 100 (percentage points)
	Significant bool    // p-value < 0.05
}

const z95 = 1.959964 // critical value for 95% two-tailed CI

// TwoProportions runs a two-proportion z-test comparing a treatment group against control.
//
// n1/conv1: control exposures/conversions
// n2/conv2: treatment exposures/conversions
//
// Returns nil when there is insufficient data to compute a valid result
// (zero exposures or zero pooled proportion denominator).
func TwoProportions(n1, conv1, n2, conv2 int64) *TwoProportionResult {
	if n1 == 0 || n2 == 0 {
		return nil
	}

	p1 := float64(conv1) / float64(n1)
	p2 := float64(conv2) / float64(n2)

	// Pooled proportion for the standard error under H0: p1 == p2.
	pPool := float64(conv1+conv2) / float64(n1+n2)
	denom := pPool * (1 - pPool) * (1/float64(n1) + 1/float64(n2))
	if denom <= 0 {
		return nil
	}

	se := math.Sqrt(denom)
	z := (p2 - p1) / se

	// Two-tailed p-value via the complementary error function.
	pValue := math.Erfc(math.Abs(z) / math.Sqrt2)

	// 95% CI uses unpooled SE (standard for interval estimation).
	seCI := math.Sqrt(p1*(1-p1)/float64(n1) + p2*(1-p2)/float64(n2))
	diff := p2 - p1
	ciLow := diff - z95*seCI
	ciHigh := diff + z95*seCI

	uplift := 0.0
	if p1 > 0 {
		uplift = diff / p1 * 100
	}

	return &TwoProportionResult{
		ZStat:       z,
		PValue:      pValue,
		CILow:       ciLow,
		CIHigh:      ciHigh,
		Uplift:      uplift,
		Significant: pValue < 0.05,
	}
}
