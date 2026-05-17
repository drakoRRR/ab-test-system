package sdk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainsdk "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/sdk"
	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
)

func TestSDKHandler_GetSdkConfig(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp sdkgen.GetSdkConfigResponseObject)
	}

	tests := []testCase{
		{
			name: "200 returns config with flags and experiments",
			ctx:  sdkCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetConfig(mock.Anything, fixedProjectID).
					Return(fixedConfig, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.GetSdkConfigResponseObject) {
				r, ok := resp.(sdkgen.GetSdkConfig200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, fixedProjectID, r.ProjectId)
				assert.Len(t, r.Flags, 1)
				assert.Equal(t, "checkout-button", r.Flags[0].Key)
				assert.True(t, r.Flags[0].Enabled)
				assert.Len(t, r.Flags[0].Rules, 1)
				assert.Len(t, r.Experiments, 1)
				assert.Equal(t, "checkout-btn-experiment", r.Experiments[0].Key)
				assert.Len(t, r.Experiments[0].Variants, 1)
			},
		},
		{
			name: "200 with empty flags and experiments",
			ctx:  sdkCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetConfig(mock.Anything, fixedProjectID).
					Return(domainsdk.Config{ProjectID: fixedProjectID}, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.GetSdkConfigResponseObject) {
				r, ok := resp.(sdkgen.GetSdkConfig200JSONResponse)
				assert.True(t, ok)
				assert.Empty(t, r.Flags)
				assert.Empty(t, r.Experiments)
			},
		},
		{
			name:      "401 on missing project context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.GetSdkConfigResponseObject) {
				assert.IsType(t, sdkgen.GetSdkConfig401JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  sdkCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetConfig(mock.Anything, fixedProjectID).
					Return(domainsdk.Config{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetSdkConfig(tc.ctx, sdkgen.GetSdkConfigRequestObject{})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
