package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

func TestService_CreateOrUpdate(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
		assertRes func(t *testing.T, got domain.User)
	}

	tests := []testCase{
		{
			name: "creates user with member role",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Upsert(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, u domain.User) (domain.User, error) { return u, nil })
			},
			assertErr: assert.NoError,
			assertRes: func(t *testing.T, got domain.User) {
				assert.Equal(t, "firebase-uid", got.FirebaseUID)
				assert.Equal(t, "user@example.com", got.Email)
				assert.Equal(t, domain.RoleMember, got.Role)
				assert.NotEmpty(t, got.ID)
			},
		},
		{
			name: "propagates storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					Upsert(mock.Anything, mock.Anything).
					Return(domain.User{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			got, err := ms.CreateOrUpdate(context.Background(), "firebase-uid", "user@example.com", "Test User", nil)

			tc.assertErr(t, err)
			if err == nil && tc.assertRes != nil {
				tc.assertRes(t, got)
			}
		})
	}
}

func TestService_GetCurrentUser(t *testing.T) {
	type testCase struct {
		name      string
		setupMock func(ms *mockedService)
		assertErr assert.ErrorAssertionFunc
	}

	fixed := domain.User{FirebaseUID: "firebase-uid", Email: "user@example.com"}

	tests := []testCase{
		{
			name: "returns user",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().GetByFirebaseUID(mock.Anything, "firebase-uid").Return(fixed, nil)
			},
			assertErr: assert.NoError,
		},
		{
			name: "wraps not found",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByFirebaseUID(mock.Anything, "firebase-uid").
					Return(domain.User{}, domain.ErrNotFound)
			},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, domain.ErrNotFound)
			},
		},
		{
			name: "wraps generic storage error",
			setupMock: func(ms *mockedService) {
				ms.storage.EXPECT().
					GetByFirebaseUID(mock.Anything, "firebase-uid").
					Return(domain.User{}, errors.New("db unavailable"))
			},
			assertErr: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			_, err := ms.GetCurrentUser(context.Background(), "firebase-uid")

			tc.assertErr(t, err)
		})
	}
}
