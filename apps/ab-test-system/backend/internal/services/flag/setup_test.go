package flag_test

import (
	"testing"

	flagsvc "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/flag"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/flag/mocks"
)

type mockedService struct {
	*flagsvc.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	storage := mocks.NewMockStorage(t)
	return &mockedService{Service: flagsvc.NewService(storage), storage: storage}
}

func ptr[T any](v T) *T { return &v }
