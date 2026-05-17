package analytics_test

import (
	"testing"

	"github.com/google/uuid"

	analyticsh "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/analytics"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/analytics/mocks"
)

var (
	fixedProjectID    = uuid.New()
	fixedExperimentID = uuid.New()
	fixedControlID    = uuid.New()
	fixedTreatmentID  = uuid.New()
)

type mockedHandler struct {
	*analyticsh.AnalyticsHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{AnalyticsHandler: analyticsh.NewHandler(svc), svc: svc}
}

func ptr[T any](v T) *T { return &v }
