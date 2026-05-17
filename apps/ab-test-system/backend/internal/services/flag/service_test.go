package flag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

func TestService_Create(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Flag)
	}

	projectID := uuid.New()

	tests := []testCase{
		{
			name: "creates flag disabled with empty rules",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, f domain.Flag) (domain.Flag, error) { return f, nil })
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Flag) {
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, projectID, got.ProjectID)
				assert.Equal(t, "checkout-button", got.Key)
				assert.Equal(t, "Checkout Button", got.Name)
				assert.False(t, got.Enabled)
				assert.Empty(t, got.Rules)
			},
		},
		{
			name: "propagates conflict error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Flag{}, domain.ErrConflict)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrConflict)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Flag{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.Create(context.Background(), projectID, "checkout-button", "Checkout Button")

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	type testCase struct {
		name      string
		newName   *string
		enabled   *bool
		rules     *[]domain.Rule
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Flag)
	}

	projectID := uuid.New()
	base := domain.Flag{
		ProjectID: projectID,
		Key:       "checkout-button",
		Name:      "Original",
		Enabled:   false,
		Rules:     []domain.Rule{},
	}

	okUpdate := func(ms *mockedService) {
		ms.storage.EXPECT().GetByKey(mock.Anything, projectID, "checkout-button").Return(base, nil)
		ms.storage.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, f domain.Flag) (domain.Flag, error) { return f, nil })
	}

	tests := []testCase{
		{
			name:      "updates name only",
			newName:   ptr("Updated"),
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Flag) {
				assert.Equal(t, "Updated", got.Name)
				assert.False(t, got.Enabled)
			},
		},
		{
			name:      "toggles enabled",
			enabled:   ptr(true),
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Flag) {
				assert.Equal(t, "Original", got.Name)
				assert.True(t, got.Enabled)
			},
		},
		{
			name:      "updates rules",
			rules:     &[]domain.Rule{{Type: "percentage", Value: 50}},
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Flag) {
				assert.Len(t, got.Rules, 1)
				assert.Equal(t, "percentage", got.Rules[0].Type)
				assert.Equal(t, float64(50), got.Rules[0].Value)
			},
		},
		{
			name:      "nil fields — nothing changes",
			setupMock: okUpdate,
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Flag) {
				assert.Equal(t, "Original", got.Name)
				assert.False(t, got.Enabled)
			},
		},
		{
			name: "not found on read",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByKey(mock.Anything, projectID, "checkout-button").
					Return(domain.Flag{}, domain.ErrNotFound)
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

			got, err := ms.Update(context.Background(), projectID, "checkout-button", tc.newName, tc.enabled, tc.rules)

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

	tests := []testCase{
		{
			name: "success",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Delete(mock.Anything, projectID, "checkout-button").Return(nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "wraps not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Delete(mock.Anything, projectID, "checkout-button").Return(domain.ErrNotFound)
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

			err := ms.Delete(context.Background(), projectID, "checkout-button")

			tc.assertErr(t, err)
		})
	}
}
