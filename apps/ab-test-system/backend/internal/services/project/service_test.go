package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
)

func TestService_Create(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Project)
	}

	orgID := uuid.New()

	tests := []testCase{
		{
			name: "creates project with generated id",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, p domain.Project) (domain.Project, error) { return p, nil })
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Project) {
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, orgID, got.OrgID)
				assert.Equal(t, "My Project", got.Name)
				assert.Equal(t, "some description", got.Description)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Project{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.Create(context.Background(), orgID, "My Project", "some description")

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got []domain.Project)
	}

	orgID := uuid.New()
	stored := []domain.Project{
		{ID: uuid.New(), OrgID: orgID, Name: "Alpha"},
		{ID: uuid.New(), OrgID: orgID, Name: "Beta"},
	}

	tests := []testCase{
		{
			name: "returns all projects",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, orgID).Return(stored, nil)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got []domain.Project) {
				assert.Len(t, got, 2)
			},
		},
		{
			name: "returns empty slice",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, orgID).Return([]domain.Project{}, nil)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got []domain.Project) {
				assert.Empty(t, got)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, orgID).Return(nil, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.List(context.Background(), orgID)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
	}

	orgID := uuid.New()
	projectID := uuid.New()
	stored := domain.Project{ID: projectID, OrgID: orgID, Name: "Alpha"}

	tests := []testCase{
		{
			name: "returns project",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().GetByID(mock.Anything, projectID, orgID).Return(stored, nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "wraps not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByID(mock.Anything, projectID, orgID).
					Return(domain.Project{}, domain.ErrNotFound)
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

			_, err := ms.GetByID(context.Background(), orgID, projectID)

			tc.assertErr(t, err)
		})
	}
}

func TestService_Update(t *testing.T) {
	type testCase struct {
		name      string
		newName   *string
		newDesc   *string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Project)
	}

	orgID := uuid.New()
	projectID := uuid.New()
	base := domain.Project{ID: projectID, OrgID: orgID, Name: "Original", Description: "Original desc"}

	okUpdate := func(ms *mockedService) {
		ms.storage.EXPECT().GetByID(mock.Anything, projectID, orgID).Return(base, nil)
		ms.storage.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, p domain.Project) (domain.Project, error) { return p, nil })
	}

	tests := []testCase{
		{
			name:      "updates both fields",
			newName:   ptr("Updated"),
			newDesc:   ptr("Updated desc"),
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Project) {
				assert.Equal(t, "Updated", got.Name)
				assert.Equal(t, "Updated desc", got.Description)
			},
		},
		{
			name:      "updates name only — description unchanged",
			newName:   ptr("Updated"),
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Project) {
				assert.Equal(t, "Updated", got.Name)
				assert.Equal(t, "Original desc", got.Description)
			},
		},
		{
			name:      "updates description only — name unchanged",
			newDesc:   ptr("Updated desc"),
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Project) {
				assert.Equal(t, "Original", got.Name)
				assert.Equal(t, "Updated desc", got.Description)
			},
		},
		{
			name:      "nil fields — nothing changes",
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Project) {
				assert.Equal(t, "Original", got.Name)
				assert.Equal(t, "Original desc", got.Description)
			},
		},
		{
			name: "not found on read",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByID(mock.Anything, projectID, orgID).
					Return(domain.Project{}, domain.ErrNotFound)
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

			got, err := ms.Update(context.Background(), orgID, projectID, tc.newName, tc.newDesc)

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

	orgID := uuid.New()
	projectID := uuid.New()

	tests := []testCase{
		{
			name: "success",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Delete(mock.Anything, projectID, orgID).Return(nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "wraps not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Delete(mock.Anything, projectID, orgID).Return(domain.ErrNotFound)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
		{
			name: "wraps generic storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Delete(mock.Anything, projectID, orgID).Return(errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			err := ms.Delete(context.Background(), orgID, projectID)

			tc.assertErr(t, err)
		})
	}
}
