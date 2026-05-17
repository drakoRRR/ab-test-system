package apikey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

func TestAPIKeyHandler_CreateApiKey(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.CreateApiKeyJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.CreateApiKeyResponseObject)
	}

	validBody := &gen.CreateApiKeyJSONRequestBody{Name: "Production"}
	fixedKey := domain.Key{ID: fixedKeyID, ProjectID: fixedProjectID, Name: "Production", Prefix: "sk_a1b2"}

	tests := []testCase{
		{
			name: "201 on success",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, fixedProjectID, "Production").
					Return(fixedKey, "sk_a1b2c3d4", nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateApiKeyResponseObject) {
				assert.IsType(t, gen.CreateApiKey201JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateApiKeyResponseObject) {
				assert.IsType(t, gen.CreateApiKey400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateApiKeyResponseObject) {
				assert.IsType(t, gen.CreateApiKey401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when project not found",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, fixedProjectID, "Production").
					Return(domain.Key{}, "", domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateApiKeyResponseObject) {
				assert.IsType(t, gen.CreateApiKey404JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.Key{}, "", errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.CreateApiKey(tc.ctx, gen.CreateApiKeyRequestObject{
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

func TestAPIKeyHandler_ListApiKeys(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.ListApiKeysResponseObject)
	}

	stored := []domain.Key{
		{ID: fixedKeyID, ProjectID: fixedProjectID, Name: "Production"},
	}

	tests := []testCase{
		{
			name: "200 returns list",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().List(mock.Anything, fixedProjectID).Return(stored, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListApiKeysResponseObject) {
				r, ok := resp.(gen.ListApiKeys200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, r, 1)
			},
		},
		{
			name: "200 returns empty list",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().List(mock.Anything, fixedProjectID).Return([]domain.Key{}, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListApiKeysResponseObject) {
				r, ok := resp.(gen.ListApiKeys200JSONResponse)
				assert.True(t, ok)
				assert.Empty(t, r)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListApiKeysResponseObject) {
				assert.IsType(t, gen.ListApiKeys401JSONResponse{}, resp)
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

			resp, err := mh.ListApiKeys(tc.ctx, gen.ListApiKeysRequestObject{
				ProjectId: fixedProjectID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestAPIKeyHandler_RevokeApiKey(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.RevokeApiKeyResponseObject)
	}

	tests := []testCase{
		{
			name: "204 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Revoke(mock.Anything, fixedProjectID, fixedKeyID).Return(nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.RevokeApiKeyResponseObject) {
				assert.IsType(t, gen.RevokeApiKey204Response{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.RevokeApiKeyResponseObject) {
				assert.IsType(t, gen.RevokeApiKey401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when key not found",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Revoke(mock.Anything, fixedProjectID, fixedKeyID).Return(domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.RevokeApiKeyResponseObject) {
				assert.IsType(t, gen.RevokeApiKey404JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Revoke(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.RevokeApiKey(tc.ctx, gen.RevokeApiKeyRequestObject{
				ProjectId: fixedProjectID,
				KeyId:     fixedKeyID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
