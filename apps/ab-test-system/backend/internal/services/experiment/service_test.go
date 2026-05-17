package experiment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

func TestService_Create(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Experiment)
	}

	projectID := uuid.New()
	variants := []domain.Variant{
		{Key: "control", Name: "Control", Weight: 50},
		{Key: "treatment", Name: "Treatment", Weight: 50},
	}

	tests := []testCase{
		{
			name: "creates experiment in draft with variants",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, exp domain.Experiment) (domain.Experiment, error) {
						return exp, nil
					})
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, projectID, got.ProjectID)
				assert.Equal(t, domain.StatusDraft, got.Status)
				assert.Equal(t, "btn-color", got.Name)
				assert.Equal(t, float64(50), got.TrafficPercent)
				assert.Len(t, got.Variants, 2)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Experiment{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.Create(context.Background(), projectID, nil, "btn-color", "", 50, variants)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	type testCase struct {
		name           string
		newName        *string
		newDescription *string
		newTraffic     *float64
		setupMock      func(ms *mockedService)
		assertErr      assert.ErrorAssertionFunc
		assertRes      func(t *testing.T, got domain.Experiment)
	}

	projectID := uuid.New()
	experimentID := uuid.New()
	draftExp := domain.Experiment{
		ID:             experimentID,
		ProjectID:      projectID,
		Name:           "Original",
		Status:         domain.StatusDraft,
		TrafficPercent: 30,
		Variants:       []domain.Variant{},
	}

	okGetUpdate := func(ms *mockedService) {
		ms.storage.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(draftExp, nil)
		ms.storage.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, exp domain.Experiment) (domain.Experiment, error) { return exp, nil })
	}

	tests := []testCase{
		{
			name:      "updates name only",
			newName:   ptr("Updated"),
			setupMock: okGetUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, "Updated", got.Name)
				assert.Equal(t, float64(30), got.TrafficPercent)
			},
		},
		{
			name:       "updates traffic percent",
			newTraffic: ptr(float64(80)),
			setupMock:  okGetUpdate,
			assertErr:  assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, "Original", got.Name)
				assert.Equal(t, float64(80), got.TrafficPercent)
			},
		},
		{
			name:      "nil fields — nothing changes",
			setupMock: okGetUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, "Original", got.Name)
				assert.Equal(t, float64(30), got.TrafficPercent)
			},
		},
		{
			name: "rejects update when not in draft",
			setupMock: func(ms *mockedService) {
				running := draftExp
				running.Status = domain.StatusRunning
				ms.storage.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(running, nil)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotDraft)
			},
		},
		{
			name: "not found on read",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByID(mock.Anything, projectID, experimentID).
					Return(domain.Experiment{}, domain.ErrNotFound)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.Update(
				context.Background(),
				projectID,
				experimentID,
				tc.newName,
				tc.newDescription,
				tc.newTraffic,
			)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
	}

	projectID := uuid.New()
	experimentID := uuid.New()
	draftExp := domain.Experiment{ID: experimentID, ProjectID: projectID, Status: domain.StatusDraft}

	tests := []testCase{
		{
			name: "deletes draft experiment",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(draftExp, nil)
				ms.storage.EXPECT().Delete(mock.Anything, projectID, experimentID).Return(nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "rejects delete when not in draft",
			setupMock: func(ms *mockedService) {
				running := draftExp
				running.Status = domain.StatusRunning
				ms.storage.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(running, nil)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotDraft)
			},
		},
		{
			name: "not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByID(mock.Anything, projectID, experimentID).
					Return(domain.Experiment{}, domain.ErrNotFound)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			err := ms.Delete(context.Background(), projectID, experimentID)
			tc.assertErr(t, err)
		})
	}
}

func TestService_Lifecycle(t *testing.T) {
	type testCase struct {
		name      string
		action    func(ms *mockedService, projectID, expID uuid.UUID) (domain.Experiment, error)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Experiment)
	}

	projectID := uuid.New()
	experimentID := uuid.New()

	makeExp := func(status domain.Status) domain.Experiment {
		return domain.Experiment{ID: experimentID, ProjectID: projectID, Status: status}
	}

	okUpdate := func(ms *mockedService, base domain.Status) {
		ms.storage.EXPECT().GetByID(mock.Anything, projectID, experimentID).Return(makeExp(base), nil)
		ms.storage.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, exp domain.Experiment) (domain.Experiment, error) { return exp, nil })
	}

	tests := []testCase{
		{
			name: "draft → running (Start)",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				okUpdate(ms, domain.StatusDraft)
				return ms.Start(context.Background(), pid, eid)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, domain.StatusRunning, got.Status)
				assert.NotNil(t, got.StartedAt)
			},
		},
		{
			name: "running → paused (Pause)",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				okUpdate(ms, domain.StatusRunning)
				return ms.Pause(context.Background(), pid, eid)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, domain.StatusPaused, got.Status)
			},
		},
		{
			name: "paused → running (Resume)",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				okUpdate(ms, domain.StatusPaused)
				return ms.Resume(context.Background(), pid, eid)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, domain.StatusRunning, got.Status)
			},
		},
		{
			name: "running → completed (Complete)",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				okUpdate(ms, domain.StatusRunning)
				return ms.Complete(context.Background(), pid, eid)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Experiment) {
				assert.Equal(t, domain.StatusCompleted, got.Status)
				assert.NotNil(t, got.EndedAt)
			},
		},
		{
			name: "completed → running is invalid",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				ms.storage.EXPECT().GetByID(mock.Anything, pid, eid).Return(makeExp(domain.StatusCompleted), nil)
				return ms.Start(context.Background(), pid, eid)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrInvalidTransition)
			},
		},
		{
			name: "draft → paused is invalid",
			action: func(ms *mockedService, pid, eid uuid.UUID) (domain.Experiment, error) {
				ms.storage.EXPECT().GetByID(mock.Anything, pid, eid).Return(makeExp(domain.StatusDraft), nil)
				return ms.Pause(context.Background(), pid, eid)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrInvalidTransition)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)

			got, err := tc.action(ms, projectID, experimentID)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
