package user_test

import (
	"context"
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/user/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

type mockedHandler struct {
	*user.UserHandler
	svc *mocks.MockService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	svc := mocks.NewMockService(t)
	return &mockedHandler{
		UserHandler: user.NewHandler(svc),
		svc:         svc,
	}
}

func authedCtx() context.Context {
	return middleware.ContextWithUserID(context.Background(), "firebase-uid")
}
