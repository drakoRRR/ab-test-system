package experiment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	experimenth "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/experiment/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

var (
	fixedProjectID    = uuid.New()
	fixedExperimentID = uuid.New()
)

type mockedHandler struct {
	*experimenth.ExperimentHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{ExperimentHandler: experimenth.NewHandler(svc), svc: svc}
}

func authedCtx() context.Context {
	return middleware.ContextWithUserID(context.Background(), "firebase-uid")
}

func ptr[T any](v T) *T { return &v }
