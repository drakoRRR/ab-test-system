package flag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

var fixedFlag = domain.Flag{
	ProjectID: fixedProjectID,
	Key:       "checkout-button",
	Name:      "Checkout Button",
	Enabled:   false,
	Rules:     []domain.Rule{},
}

func TestFlagHandler_CreateFlag(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.CreateFlagJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.CreateFlagResponseObject)
	}

	validBody := &gen.CreateFlagJSONRequestBody{Key: "checkout-button", Name: "Checkout Button"}

	tests := []testCase{
		{
			name: "201 on success",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, fixedProjectID, "checkout-button", "Checkout Button").
					Return(fixedFlag, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateFlagResponseObject) {
				assert.IsType(t, gen.CreateFlag201JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateFlagResponseObject) {
				assert.IsType(t, gen.CreateFlag400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateFlagResponseObject) {
				assert.IsType(t, gen.CreateFlag401JSONResponse{}, resp)
			},
		},
		{
			name: "409 on duplicate key",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.Flag{}, domain.ErrConflict)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateFlagResponseObject) {
				assert.IsType(t, gen.CreateFlag409JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.Flag{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.CreateFlag(tc.ctx, gen.CreateFlagRequestObject{
				ProjectId: fixedProjectID,
				Body:      tc.body,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestFlagHandler_ListFlags(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.ListFlagsResponseObject)
	}

	tests := []testCase{
		{
			name: "200 returns list",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().List(mock.Anything, fixedProjectID).
					Return([]domain.Flag{fixedFlag}, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListFlagsResponseObject) {
				r, ok := resp.(gen.ListFlags200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, r, 1)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListFlagsResponseObject) {
				assert.IsType(t, gen.ListFlags401JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.ListFlags(tc.ctx, gen.ListFlagsRequestObject{
				ProjectId: fixedProjectID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestFlagHandler_GetFlag(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.GetFlagResponseObject)
	}

	tests := []testCase{
		{
			name: "200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().GetByKey(mock.Anything, fixedProjectID, "checkout-button").Return(fixedFlag, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetFlagResponseObject) {
				assert.IsType(t, gen.GetFlag200JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetFlagResponseObject) {
				assert.IsType(t, gen.GetFlag401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetByKey(mock.Anything, fixedProjectID, "checkout-button").
					Return(domain.Flag{}, domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetFlagResponseObject) {
				assert.IsType(t, gen.GetFlag404JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetFlag(tc.ctx, gen.GetFlagRequestObject{
				ProjectId: fixedProjectID,
				FlagKey:   "checkout-button",
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestFlagHandler_UpdateFlag(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.UpdateFlagJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.UpdateFlagResponseObject)
	}

	tests := []testCase{
		{
			name: "200 toggle enabled",
			ctx:  authedCtx(),
			body: &gen.UpdateFlagJSONRequestBody{Enabled: ptr(true)},
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Update(mock.Anything, domain.UpdateParams{
						ProjectID: fixedProjectID,
						Key:       "checkout-button",
						Enabled:   ptr(true),
					}).
					Return(fixedFlag, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateFlagResponseObject) {
				assert.IsType(t, gen.UpdateFlag200JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateFlagResponseObject) {
				assert.IsType(t, gen.UpdateFlag400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      &gen.UpdateFlagJSONRequestBody{Name: ptr("New Name")},
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateFlagResponseObject) {
				assert.IsType(t, gen.UpdateFlag401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			body: &gen.UpdateFlagJSONRequestBody{Name: ptr("New Name")},
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(domain.Flag{}, domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateFlagResponseObject) {
				assert.IsType(t, gen.UpdateFlag404JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.UpdateFlag(tc.ctx, gen.UpdateFlagRequestObject{
				ProjectId: fixedProjectID,
				FlagKey:   "checkout-button",
				Body:      tc.body,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestFlagHandler_DeleteFlag(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.DeleteFlagResponseObject)
	}

	tests := []testCase{
		{
			name: "204 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Delete(mock.Anything, fixedProjectID, "checkout-button").Return(nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteFlagResponseObject) {
				assert.IsType(t, gen.DeleteFlag204Response{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteFlagResponseObject) {
				assert.IsType(t, gen.DeleteFlag401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Delete(mock.Anything, fixedProjectID, "checkout-button").Return(domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteFlagResponseObject) {
				assert.IsType(t, gen.DeleteFlag404JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.DeleteFlag(tc.ctx, gen.DeleteFlagRequestObject{
				ProjectId: fixedProjectID,
				FlagKey:   "checkout-button",
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
