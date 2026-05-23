package experiment_test

import (
	"context"
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment"
	"github.com/google/uuid"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

var (
	fixedProjectID    = uuid.New()
	fixedExperimentID = uuid.New()
)

type mockedHandler struct {
	*experiment.ExperimentHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{ExperimentHandler: experiment.NewHandler(svc), svc: svc}
}

func authedCtx() context.Context {
	return middleware.ContextWithUserID(context.Background(), "firebase-uid")
}

func ptr[T any](v T) *T { return &v }
