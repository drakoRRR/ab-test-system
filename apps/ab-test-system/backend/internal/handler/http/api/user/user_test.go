package user_test

import (
	"context"
	"errors"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func TestUserHandler_CreateUser(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.CreateUserJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.CreateUserResponseObject)
	}

	validBody := &gen.CreateUserJSONRequestBody{
		Email: openapi_types.Email("user@example.com"),
		Name:  "Test User",
	}

	fixedUser := domain.User{FirebaseUID: "firebase-uid", Email: "user@example.com", Role: domain.RoleMember}

	tests := []testCase{
		{
			name: "201 on success",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					CreateOrUpdate(mock.Anything, "firebase-uid", string(validBody.Email), validBody.Name, mock.Anything).
					Return(fixedUser, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateUserResponseObject) {
				assert.IsType(t, gen.CreateUser201JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateUserResponseObject) {
				assert.IsType(t, gen.CreateUser400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateUserResponseObject) {
				assert.IsType(t, gen.CreateUser401JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					CreateOrUpdate(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.CreateUser(tc.ctx, gen.CreateUserRequestObject{Body: tc.body})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestUserHandler_GetCurrentUser(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.GetCurrentUserResponseObject)
	}

	fixedUser := domain.User{FirebaseUID: "firebase-uid", Email: "user@example.com", Role: domain.RoleMember}

	tests := []testCase{
		{
			name: "200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().GetCurrentUser(mock.Anything, "firebase-uid").Return(fixedUser, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetCurrentUserResponseObject) {
				assert.IsType(t, gen.GetCurrentUser200JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetCurrentUserResponseObject) {
				assert.IsType(t, gen.GetCurrentUser401JSONResponse{}, resp)
			},
		},
		{
			name: "401 when user not registered",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().GetCurrentUser(mock.Anything, "firebase-uid").Return(domain.User{}, domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetCurrentUserResponseObject) {
				assert.IsType(t, gen.GetCurrentUser401JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetCurrentUser(mock.Anything, "firebase-uid").
					Return(domain.User{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetCurrentUser(tc.ctx, gen.GetCurrentUserRequestObject{})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
