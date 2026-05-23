package apikey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
)

func TestService_Create(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.Key, raw string)
	}

	projectID := uuid.New()

	tests := []testCase{
		{
			name: "creates key with prefix and hash",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, k domain.Key) (domain.Key, error) { return k, nil })
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.Key, raw string) {
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, projectID, got.ProjectID)
				assert.Equal(t, "Production", got.Name)
				assert.NotEmpty(t, got.KeyHash)
				assert.NotEqual(t, raw, got.KeyHash)
				assert.Equal(t, raw[:7], got.Prefix)
				assert.True(t, len(raw) > 7)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Key{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, raw, err := ms.Create(context.Background(), projectID, "Production")

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got, raw)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got []domain.Key)
	}

	projectID := uuid.New()
	stored := []domain.Key{
		{ID: uuid.New(), ProjectID: projectID, Name: "Production"},
		{ID: uuid.New(), ProjectID: projectID, Name: "Staging"},
	}

	tests := []testCase{
		{
			name: "returns all keys",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, projectID).Return(stored, nil)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got []domain.Key) {
				assert.Len(t, got, 2)
			},
		},
		{
			name: "returns empty slice",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, projectID).Return([]domain.Key{}, nil)
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got []domain.Key) {
				assert.Empty(t, got)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().List(mock.Anything, projectID).Return(nil, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.List(context.Background(), projectID)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_Revoke(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
	}

	projectID := uuid.New()
	keyID := uuid.New()

	tests := []testCase{
		{
			name: "success",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Revoke(mock.Anything, keyID, projectID).Return(nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "wraps not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Revoke(mock.Anything, keyID, projectID).Return(domain.ErrNotFound)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().Revoke(mock.Anything, keyID, projectID).Return(errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			err := ms.Revoke(context.Background(), projectID, keyID)

			tc.assertErr(t, err)
		})
	}
}
