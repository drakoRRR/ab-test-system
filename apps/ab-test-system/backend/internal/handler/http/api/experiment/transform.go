package experiment

import (
	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIExperiment(exp domain.Experiment) gen.Experiment {
	variants := make([]gen.Variant, len(exp.Variants))
	for i, v := range exp.Variants {
		variants[i] = gen.Variant{
			Id:     v.ID,
			Key:    v.Key,
			Name:   v.Name,
			Weight: v.Weight,
		}
	}

	out := gen.Experiment{
		Id:             exp.ID,
		ProjectId:      exp.ProjectID,
		Key:            exp.Key,
		Name:           exp.Name,
		Status:         gen.ExperimentStatus(exp.Status),
		TrafficPercent: float32(exp.TrafficPercent),
		Variants:       variants,
		CreatedAt:      exp.CreatedAt,
		UpdatedAt:      exp.UpdatedAt,
		StartedAt:      exp.StartedAt,
		EndedAt:        exp.EndedAt,
	}

	if exp.Description != "" {
		out.Description = &exp.Description
	}

	if exp.FlagID != nil {
		out.FlagId = exp.FlagID
	}

	return out
}

func toDomainVariants(variants []gen.CreateVariantRequest) []domain.Variant {
	out := make([]domain.Variant, len(variants))
	for i, v := range variants {
		out[i] = domain.Variant{
			Key:    v.Key,
			Name:   v.Name,
			Weight: v.Weight,
		}
	}

	return out
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func ptr[T any](v T) *T { return &v }
