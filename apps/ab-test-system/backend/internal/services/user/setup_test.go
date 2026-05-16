package user_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/user"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/user/mocks"
)

type mockedService struct {
	*user.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	storage := mocks.NewMockStorage(t)
	return &mockedService{
		Service: user.NewService(storage),
		storage: storage,
	}
}
