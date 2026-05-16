package project_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	domainproject "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
	domainuser "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/project/mocks"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/middleware"
)

var (
	fixedOrgID     = uuid.New()
	fixedProjectID = uuid.New()

	fixedUser = domainuser.User{
		FirebaseUID: "firebase-uid",
		OrgID:       &fixedOrgID,
	}

	fixedUserNoOrg = domainuser.User{
		FirebaseUID: "firebase-uid",
		OrgID:       nil,
	}

	fixedProject = domainproject.Project{
		ID:    fixedProjectID,
		OrgID: fixedOrgID,
		Name:  "My Project",
	}
)

type mockedHandler struct {
	*project.ProjectHandler
	userSvc *mocks.MockUserService
	projSvc *mocks.MockProjectService
}

func newMockedHandler(t *testing.T) *mockedHandler {
	userSvc := mocks.NewMockUserService(t)
	projSvc := mocks.NewMockProjectService(t)
	return &mockedHandler{
		ProjectHandler: project.NewHandler(userSvc, projSvc),
		userSvc:        userSvc,
		projSvc:        projSvc,
	}
}

func authedCtx() context.Context {
	return middleware.ContextWithUserID(context.Background(), "firebase-uid")
}

func withUserOK(mh *mockedHandler) {
	mh.userSvc.EXPECT().GetCurrentUser(mock.Anything, "firebase-uid").Return(fixedUser, nil)
}

func withUserNoOrg(mh *mockedHandler) {
	mh.userSvc.EXPECT().GetCurrentUser(mock.Anything, "firebase-uid").Return(fixedUserNoOrg, nil)
}

func ptr[T any](v T) *T { return &v }
