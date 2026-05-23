package project_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/project"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/project/mocks"
)

type mockedService struct {
	*project.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	storage := mocks.NewMockStorage(t)
	return &mockedService{
		Service: project.NewService(storage),
		storage: storage,
	}
}

func ptr[T any](v T) *T { return &v }
