package analytics

import "github.com/google/uuid"

type VariantMetrics struct {
	VariantID   uuid.UUID
	Exposures   int64
	Conversions int64
}

type VariantResult struct {
	VariantID      uuid.UUID
	VariantKey     string
	VariantName    string
	Exposures      int64
	Conversions    int64
	ConversionRate float64

	// Statistical fields — nil for the control variant.
	IsControl   bool
	Uplift      *float64 // percentage points relative to control, e.g. 3.5 means +3.5%
	PValue      *float64
	CILow       *float64
	CIHigh      *float64
	Significant *bool // true when p-value < 0.05
}

type ExperimentResult struct {
	ExperimentID     uuid.UUID
	TotalExposures   int64
	TotalConversions int64
	Variants         []VariantResult
}
