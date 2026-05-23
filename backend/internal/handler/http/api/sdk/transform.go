package sdk

import (
	"github.com/google/uuid"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
)

func toAPIConfig(cfg domainsdk.Config) sdkgen.SDKConfig {
	return sdkgen.SDKConfig{
		ProjectId:   cfg.ProjectID,
		Flags:       toAPIFlags(cfg.Flags),
		Experiments: toAPIExperiments(cfg.Experiments),
	}
}

func toAPIFlags(flags []domainflag.Flag) []sdkgen.SDKFlag {
	out := make([]sdkgen.SDKFlag, len(flags))
	for i, f := range flags {
		out[i] = sdkgen.SDKFlag{
			Id:      f.ID,
			Key:     f.Key,
			Enabled: f.Enabled,
			Rules:   toAPIFlagRules(f.Rules),
		}
	}

	return out
}

func toAPIFlagRules(rules []domainflag.Rule) []sdkgen.SDKFlagRule {
	out := make([]sdkgen.SDKFlagRule, len(rules))
	for i, r := range rules {
		out[i] = sdkgen.SDKFlagRule{
			Type:  sdkgen.SDKFlagRuleType(r.Type),
			Value: float32(r.Value),
		}
	}

	return out
}

func toAPIExperiments(experiments []domainexp.Experiment) []sdkgen.SDKExperiment {
	out := make([]sdkgen.SDKExperiment, len(experiments))
	for i, e := range experiments {
		out[i] = sdkgen.SDKExperiment{
			Id:             e.ID,
			Key:            e.Key,
			TrafficPercent: float32(e.TrafficPercent),
			Variants:       toAPIVariants(e.Variants),
		}
	}

	return out
}

func toAPIVariants(variants []domainexp.Variant) []sdkgen.SDKVariant {
	out := make([]sdkgen.SDKVariant, len(variants))
	for i, v := range variants {
		out[i] = sdkgen.SDKVariant{
			Id:     v.ID,
			Key:    v.Key,
			Weight: v.Weight,
		}
	}

	return out
}

func toDomainEvents(projectID uuid.UUID, batch []sdkgen.SDKEvent) []domainevent.Event {
	out := make([]domainevent.Event, len(batch))
	for i, e := range batch {
		out[i] = domainevent.Event{
			ID:           e.Id,
			ProjectID:    projectID,
			UserID:       e.UserId,
			ExperimentID: e.ExperimentId,
			VariantID:    e.VariantId,
			Type:         domainevent.Type(e.Type),
			Name:         derefString(e.Name),
			Value:        float64(derefFloat32(e.Value)),
			Timestamp:    e.Timestamp,
		}
	}
	return out
}

func ptr[T any](v T) *T { return &v }

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefFloat32(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}
