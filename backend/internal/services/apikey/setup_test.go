package apikey_test

import (
	"testing"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/apikey"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/apikey/mocks"
)

type mockedService struct {
	*apikey.Service
	storage *mocks.MockStorage
}

func newMockedService(t *testing.T) *mockedService {
	storage := mocks.NewMockStorage(t)
	return &mockedService{Service: apikey.NewService(storage), storage: storage}
}
