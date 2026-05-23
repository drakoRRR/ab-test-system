package sdk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sdkgen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen/sdk"
)

func TestSDKHandler_PostSdkEvents(t *testing.T) {
	validBatch := &sdkgen.SDKEventBatch{
		Events: []sdkgen.SDKEvent{
			{
				Id:           fixedFlagID,
				UserId:       "user-1",
				ExperimentId: fixedExperimentID,
				VariantId:    fixedVariantID,
				Type:         sdkgen.Exposure,
				Timestamp:    time.Now().UTC(),
			},
		},
	}

	type testCase struct {
		name       string
		ctx        context.Context
		body       *sdkgen.SDKEventBatch
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp sdkgen.PostSdkEventsResponseObject)
	}

	tests := []testCase{
		{
			name: "202 on valid batch",
			ctx:  sdkCtx(),
			body: validBatch,
			setupMock: func(mh *mockedHandler) {
				mh.events.EXPECT().
					Ingest(mock.Anything, mock.Anything).
					Return(nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.PostSdkEventsResponseObject) {
				assert.IsType(t, sdkgen.PostSdkEvents202Response{}, resp)
			},
		},
		{
			name:      "401 on missing project context",
			ctx:       context.Background(),
			body:      validBatch,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.PostSdkEventsResponseObject) {
				assert.IsType(t, sdkgen.PostSdkEvents401JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       sdkCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp sdkgen.PostSdkEventsResponseObject) {
				assert.IsType(t, sdkgen.PostSdkEvents400JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  sdkCtx(),
			body: validBatch,
			setupMock: func(mh *mockedHandler) {
				mh.events.EXPECT().
					Ingest(mock.Anything, mock.Anything).
					Return(errors.New("kafka unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.PostSdkEvents(tc.ctx, sdkgen.PostSdkEventsRequestObject{Body: tc.body})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
