package analytics_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	domainexperiment "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	analyticssvc "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/analytics"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/analytics/mocks"
)

var (
	projectID    = uuid.New()
	experimentID = uuid.New()
	controlID    = uuid.New()
	treatmentID  = uuid.New()

	baseExperiment = domainexperiment.Experiment{
		ID:             experimentID,
		ProjectID:      projectID,
		Key:            "btn-color",
		Name:           "Button color test",
		Status:         domainexperiment.StatusRunning,
		TrafficPercent: 100,
		Variants: []domainexperiment.Variant{
			{ID: controlID, Key: "control", Name: "Control", Weight: 50},
			{ID: treatmentID, Key: "treatment", Name: "Treatment", Weight: 50},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
)

type mockedService struct {
	*analyticssvc.Service
	metrics     *mocks.MockMetricsStorage
	experiments *mocks.MockExperimentStorage
}

func newMockedService(t *testing.T) *mockedService {
	t.Helper()
	metrics := mocks.NewMockMetricsStorage(t)
	experiments := mocks.NewMockExperimentStorage(t)
	return &mockedService{
		Service:     analyticssvc.NewService(metrics, experiments),
		metrics:     metrics,
		experiments: experiments,
	}
}

func errIs(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, _ ...interface{}) bool {
		return assert.ErrorIs(t, err, target)
	}
}
