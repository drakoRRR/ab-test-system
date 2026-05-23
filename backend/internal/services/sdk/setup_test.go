package sdk_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/sdk"
	"github.com/google/uuid"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/sdk/mocks"

	domainexp "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	domainflag "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

var (
	projectID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	flagID       = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	experimentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	variantID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

var (
	activeFlag = domainflag.Flag{
		ID:        flagID,
		ProjectID: projectID,
		Key:       "checkout-button",
		Enabled:   true,
		Rules:     []domainflag.Rule{{Type: "percentage", Value: 50}},
	}

	runningExperiment = domainexp.Experiment{
		ID:             experimentID,
		ProjectID:      projectID,
		Key:            "checkout-btn-experiment",
		Status:         domainexp.StatusRunning,
		TrafficPercent: 50,
		Variants:       []domainexp.Variant{{ID: variantID, Key: "control", Weight: 50}},
	}

	draftExperiment = domainexp.Experiment{
		ID:        uuid.New(),
		ProjectID: projectID,
		Key:       "draft-experiment",
		Status:    domainexp.StatusDraft,
	}

	pausedExperiment = domainexp.Experiment{
		ID:        uuid.New(),
		ProjectID: projectID,
		Key:       "paused-experiment",
		Status:    domainexp.StatusPaused,
	}
)

type mockedService struct {
	*sdk.Service
	flags       *mocks.MockFlagLister
	experiments *mocks.MockExperimentLister
}

func newMockedService(t *testing.T) *mockedService {
	flags := mocks.NewMockFlagLister(t)
	experiments := mocks.NewMockExperimentLister(t)
	return &mockedService{
		Service:     sdk.NewService(flags, experiments),
		flags:       flags,
		experiments: experiments,
	}
}
