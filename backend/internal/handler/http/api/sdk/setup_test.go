package sdk_test

import (
	"context"
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/sdk"
	"github.com/google/uuid"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/sdk/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"

	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
)

var (
	fixedProjectID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedFlagID       = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	fixedExperimentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	fixedVariantID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

var fixedConfig = domainsdk.Config{
	ProjectID: fixedProjectID,
	Flags: []domainflag.Flag{
		{
			ID:        fixedFlagID,
			ProjectID: fixedProjectID,
			Key:       "checkout-button",
			Enabled:   true,
			Rules:     []domainflag.Rule{{Type: "percentage", Value: 50}},
		},
	},
	Experiments: []domainexp.Experiment{
		{
			ID:             fixedExperimentID,
			ProjectID:      fixedProjectID,
			Key:            "checkout-btn-experiment",
			Status:         domainexp.StatusRunning,
			TrafficPercent: 50,
			Variants: []domainexp.Variant{
				{ID: fixedVariantID, Key: "control", Name: "Control", Weight: 50},
			},
		},
	},
}

type mockedHandler struct {
	*sdk.Handler
	config *mocks.MockConfigService
	events *mocks.MockEventService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	config := mocks.NewMockConfigService(t)
	events := mocks.NewMockEventService(t)
	return &mockedHandler{
		Handler: sdk.NewHandler(config, events),
		config:  config,
		events:  events,
	}
}

func sdkCtx() context.Context {
	return middleware.ContextWithProjectID(context.Background(), fixedProjectID)
}
