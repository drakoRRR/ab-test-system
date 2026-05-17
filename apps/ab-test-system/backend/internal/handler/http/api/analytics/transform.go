package analytics

import (
	domainanalytics "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/analytics"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIAnalytics(r domainanalytics.ExperimentResult) gen.ExperimentAnalytics {
	variants := make([]gen.VariantAnalytics, 0, len(r.Variants))
	for _, v := range r.Variants {
		variants = append(variants, toAPIVariant(v))
	}
	return gen.ExperimentAnalytics{
		ExperimentId:     r.ExperimentID,
		TotalExposures:   r.TotalExposures,
		TotalConversions: r.TotalConversions,
		Variants:         variants,
	}
}

func toAPIVariant(v domainanalytics.VariantResult) gen.VariantAnalytics {
	return gen.VariantAnalytics{
		VariantId:      v.VariantID,
		VariantKey:     v.VariantKey,
		VariantName:    v.VariantName,
		Exposures:      v.Exposures,
		Conversions:    v.Conversions,
		ConversionRate: v.ConversionRate,
		IsControl:      v.IsControl,
		Uplift:         v.Uplift,
		PValue:         v.PValue,
		CiLow:          v.CILow,
		CiHigh:         v.CIHigh,
		Significant:    v.Significant,
	}
}
