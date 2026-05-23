package apikey_test

import (
	"context"
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/apikey"
	"github.com/google/uuid"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/apikey/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

var (
	fixedProjectID = uuid.New()
	fixedKeyID     = uuid.New()
)

type mockedHandler struct {
	*apikey.APIKeyHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{APIKeyHandler: apikey.NewHandler(svc), svc: svc}
}

func authedCtx() context.Context {
	return middleware.ContextWithUserID(context.Background(), "firebase-uid")
}
