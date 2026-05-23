package experiment_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/experiment/mocks"
)

type mockedService struct {
	*experiment.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	t.Helper()

	storage := mocks.NewMockStorage(t)

	return &mockedService{
		Service: experiment.NewService(storage),
		storage: storage,
	}
}
