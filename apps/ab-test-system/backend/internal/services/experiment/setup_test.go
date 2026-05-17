package experiment_test

import (
	"testing"

	svc "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/experiment/mocks"
)

type mockedService struct {
	*svc.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	t.Helper()

	storage := mocks.NewMockStorage(t)

	return &mockedService{
		Service: svc.NewService(storage),
		storage: storage,
	}
}
