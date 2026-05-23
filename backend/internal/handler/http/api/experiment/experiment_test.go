package experiment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	gen "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/handler/http/api/codegen"
)

var fixedExperiment = domain.Experiment{
	ID:             fixedExperimentID,
	ProjectID:      fixedProjectID,
	Key:            "checkout-btn-experiment",
	Name:           "Checkout Button Experiment",
	Status:         domain.StatusDraft,
	TrafficPercent: 50,
	Variants: []domain.Variant{
		{Key: "control", Name: "Control", Weight: 50},
		{Key: "treatment", Name: "Treatment", Weight: 50},
	},
}

func TestExperimentHandler_CreateExperiment(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.CreateExperimentJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.CreateExperimentResponseObject)
	}

	validBody := &gen.CreateExperimentJSONRequestBody{
		Key:            "checkout-btn-experiment",
		Name:           "Checkout Button Experiment",
		TrafficPercent: 50,
		Variants: []gen.CreateVariantRequest{
			{Key: "control", Name: "Control", Weight: 50},
			{Key: "treatment", Name: "Treatment", Weight: 50},
		},
	}

	tests := []testCase{
		{
			name: "201 on success",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, domain.CreateParams{
						ProjectID:      fixedProjectID,
						Key:            "checkout-btn-experiment",
						Name:           "Checkout Button Experiment",
						TrafficPercent: 50,
						Variants: []domain.Variant{
							{Key: "control", Name: "Control", Weight: 50},
							{Key: "treatment", Name: "Treatment", Weight: 50},
						},
					}).
					Return(fixedExperiment, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateExperimentResponseObject) {
				assert.IsType(t, gen.CreateExperiment201JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateExperimentResponseObject) {
				assert.IsType(t, gen.CreateExperiment400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      validBody,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateExperimentResponseObject) {
				assert.IsType(t, gen.CreateExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "409 on conflict",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Experiment{}, domain.ErrConflict)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.CreateExperimentResponseObject) {
				assert.IsType(t, gen.CreateExperiment409JSONResponse{}, resp)
			},
		},
		{
			name: "service error bubbles up as 500",
			ctx:  authedCtx(),
			body: validBody,
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Experiment{}, errors.New("db error"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.CreateExperiment(tc.ctx, gen.CreateExperimentRequestObject{
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

func TestExperimentHandler_ListExperiments(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.ListExperimentsResponseObject)
	}

	tests := []testCase{
		{
			name: "200 returns list",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().List(mock.Anything, fixedProjectID).
					Return([]domain.Experiment{fixedExperiment}, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListExperimentsResponseObject) {
				r, ok := resp.(gen.ListExperiments200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, r, 1)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.ListExperimentsResponseObject) {
				assert.IsType(t, gen.ListExperiments401JSONResponse{}, resp)
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

			resp, err := mh.ListExperiments(tc.ctx, gen.ListExperimentsRequestObject{
				ProjectId: fixedProjectID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestExperimentHandler_GetExperiment(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.GetExperimentResponseObject)
	}

	tests := []testCase{
		{
			name: "200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().GetByID(mock.Anything, fixedProjectID, fixedExperimentID).Return(fixedExperiment, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetExperimentResponseObject) {
				assert.IsType(t, gen.GetExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetExperimentResponseObject) {
				assert.IsType(t, gen.GetExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					GetByID(mock.Anything, fixedProjectID, fixedExperimentID).
					Return(domain.Experiment{}, domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.GetExperimentResponseObject) {
				assert.IsType(t, gen.GetExperiment404JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.GetExperiment(tc.ctx, gen.GetExperimentRequestObject{
				ProjectId:    fixedProjectID,
				ExperimentId: fixedExperimentID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestExperimentHandler_UpdateExperiment(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		body       *gen.UpdateExperimentJSONRequestBody
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.UpdateExperimentResponseObject)
	}

	tests := []testCase{
		{
			name: "200 updates name",
			ctx:  authedCtx(),
			body: &gen.UpdateExperimentJSONRequestBody{Name: ptr("New Name")},
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Update(mock.Anything, domain.UpdateParams{
						ProjectID:    fixedProjectID,
						ExperimentID: fixedExperimentID,
						Name:         ptr("New Name"),
					}).
					Return(fixedExperiment, nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateExperimentResponseObject) {
				assert.IsType(t, gen.UpdateExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "400 on nil body",
			ctx:       authedCtx(),
			body:      nil,
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateExperimentResponseObject) {
				assert.IsType(t, gen.UpdateExperiment400JSONResponse{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			body:      &gen.UpdateExperimentJSONRequestBody{Name: ptr("X")},
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateExperimentResponseObject) {
				assert.IsType(t, gen.UpdateExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			body: &gen.UpdateExperimentJSONRequestBody{Name: ptr("X")},
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(domain.Experiment{}, domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateExperimentResponseObject) {
				assert.IsType(t, gen.UpdateExperiment404JSONResponse{}, resp)
			},
		},
		{
			name: "409 when not in draft",
			ctx:  authedCtx(),
			body: &gen.UpdateExperimentJSONRequestBody{Name: ptr("X")},
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(domain.Experiment{}, domain.ErrNotDraft)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.UpdateExperimentResponseObject) {
				assert.IsType(t, gen.UpdateExperiment409JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.UpdateExperiment(tc.ctx, gen.UpdateExperimentRequestObject{
				ProjectId:    fixedProjectID,
				ExperimentId: fixedExperimentID,
				Body:         tc.body,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestExperimentHandler_DeleteExperiment(t *testing.T) {
	type testCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp gen.DeleteExperimentResponseObject)
	}

	tests := []testCase{
		{
			name: "204 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Delete(mock.Anything, fixedProjectID, fixedExperimentID).Return(nil)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteExperimentResponseObject) {
				assert.IsType(t, gen.DeleteExperiment204Response{}, resp)
			},
		},
		{
			name:      "401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteExperimentResponseObject) {
				assert.IsType(t, gen.DeleteExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "404 when not found",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Delete(mock.Anything, fixedProjectID, fixedExperimentID).Return(domain.ErrNotFound)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteExperimentResponseObject) {
				assert.IsType(t, gen.DeleteExperiment404JSONResponse{}, resp)
			},
		},
		{
			name: "409 when not in draft",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Delete(mock.Anything, fixedProjectID, fixedExperimentID).Return(domain.ErrNotDraft)
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp gen.DeleteExperimentResponseObject) {
				assert.IsType(t, gen.DeleteExperiment409JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := mh.DeleteExperiment(tc.ctx, gen.DeleteExperimentRequestObject{
				ProjectId:    fixedProjectID,
				ExperimentId: fixedExperimentID,
			})

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestExperimentHandler_Lifecycle(t *testing.T) {
	type lifecycleCase struct {
		name       string
		ctx        context.Context
		setupMock  func(mh *mockedHandler)
		do         func(mh *mockedHandler, ctx context.Context) (interface{}, error)
		assertErr  assert.ErrorAssertionFunc
		assertResp func(t *testing.T, resp interface{})
	}

	tests := []lifecycleCase{
		{
			name: "Start — 200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Start(mock.Anything, fixedProjectID, fixedExperimentID).Return(fixedExperiment, nil)
			},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.StartExperiment(ctx, gen.StartExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.StartExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "Start — 401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.StartExperiment(ctx, gen.StartExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.StartExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "Start — 409 on invalid transition",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().
					Start(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.Experiment{}, domain.ErrInvalidTransition)
			},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.StartExperiment(ctx, gen.StartExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.StartExperiment409JSONResponse{}, resp)
			},
		},
		{
			name: "Pause — 200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Pause(mock.Anything, fixedProjectID, fixedExperimentID).Return(fixedExperiment, nil)
			},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.PauseExperiment(ctx, gen.PauseExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.PauseExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "Pause — 401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.PauseExperiment(ctx, gen.PauseExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.PauseExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "Resume — 200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Resume(mock.Anything, fixedProjectID, fixedExperimentID).Return(fixedExperiment, nil)
			},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.ResumeExperiment(ctx, gen.ResumeExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.ResumeExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "Resume — 401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.ResumeExperiment(ctx, gen.ResumeExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.ResumeExperiment401JSONResponse{}, resp)
			},
		},
		{
			name: "Complete — 200 on success",
			ctx:  authedCtx(),
			setupMock: func(mh *mockedHandler) {
				mh.svc.EXPECT().Complete(mock.Anything, fixedProjectID, fixedExperimentID).Return(fixedExperiment, nil)
			},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.CompleteExperiment(ctx, gen.CompleteExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.CompleteExperiment200JSONResponse{}, resp)
			},
		},
		{
			name:      "Complete — 401 on missing auth context",
			ctx:       context.Background(),
			setupMock: func(_ *mockedHandler) {},
			do: func(mh *mockedHandler, ctx context.Context) (interface{}, error) {
				return mh.CompleteExperiment(ctx, gen.CompleteExperimentRequestObject{
					ProjectId: fixedProjectID, ExperimentId: fixedExperimentID,
				})
			},
			assertErr: assert.NoError,
			assertResp: func(t *testing.T, resp interface{}) {
				assert.IsType(t, gen.CompleteExperiment401JSONResponse{}, resp)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mh := newMockedHandler(t)
			tc.setupMock(mh)

			resp, err := tc.do(mh, tc.ctx)

			tc.assertErr(t, err)
			if err == nil && tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}
