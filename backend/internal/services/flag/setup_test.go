package flag_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/flag"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/flag/mocks"
)

type mockedService struct {
	*flag.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	storage := mocks.NewMockStorage(t)
	return &mockedService{Service: flag.NewService(storage), storage: storage}
}

func ptr[T any](v T) *T { return &v }
