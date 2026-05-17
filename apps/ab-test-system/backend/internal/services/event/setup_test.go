package event_test

import (
	"testing"

	eventsvc "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/event"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/services/event/mocks"
)

type mockedService struct {
	*eventsvc.Service
	publisher *mocks.MockPublisher
}

func newMockedService(t *testing.T) *mockedService {
	t.Helper()
	publisher := mocks.NewMockPublisher(t)
	return &mockedService{Service: eventsvc.NewService(publisher), publisher: publisher}
}
