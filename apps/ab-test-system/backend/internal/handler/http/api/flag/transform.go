package flag

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func toAPIFlag(f domain.Flag) gen.Flag {
	rules := make([]gen.FlagRule, len(f.Rules))
	for i, r := range f.Rules {
		rules[i] = gen.FlagRule{
			Type:  gen.FlagRuleType(r.Type),
			Value: float32(r.Value),
		}
	}

	return gen.Flag{
		Id:        openapi_types.UUID(f.ID),
		ProjectId: openapi_types.UUID(f.ProjectID),
		Key:       f.Key,
		Name:      f.Name,
		Enabled:   f.Enabled,
		Rules:     &rules,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func toDomainRules(rules []gen.FlagRule) []domain.Rule {
	out := make([]domain.Rule, len(rules))
	for i, r := range rules {
		out[i] = domain.Rule{
			Type:  string(r.Type),
			Value: float64(r.Value),
		}
	}

	return out
}

func ptr[T any](v T) *T { return &v }
