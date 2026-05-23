package analytics_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/analytics"
	"github.com/google/uuid"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/analytics/mocks"
)

var (
	fixedProjectID    = uuid.New()
	fixedExperimentID = uuid.New()
	fixedControlID    = uuid.New()
	fixedTreatmentID  = uuid.New()
)

type mockedHandler struct {
	*analytics.AnalyticsHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{AnalyticsHandler: analytics.NewHandler(svc), svc: svc}
}

func ptr[T any](v T) *T { return &v }
