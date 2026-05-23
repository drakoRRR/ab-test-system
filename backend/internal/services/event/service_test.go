package event_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
)

var fixedEvents = []domainevent.Event{
	{
		ID:           uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		ProjectID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID:       "user-1",
		ExperimentID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		VariantID:    uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		Type:         domainevent.TypeExposure,
		Timestamp:    time.Now().UTC(),
	},
}

func TestService_Ingest(t *testing.T) {
	tests := []struct {
		name      string
		events    []domainevent.Event
		setupMock func(ms *mockedService)
		wantErr   bool
	}{
		{
			name:   "publishes events to publisher",
			events: fixedEvents,
			setupMock: func(ms *mockedService) {
				ms.publisher.EXPECT().
					Publish(mock.Anything, fixedEvents).
					Return(nil)
			},
		},
		{
			name:   "empty batch still calls publisher",
			events: []domainevent.Event{},
			setupMock: func(ms *mockedService) {
				ms.publisher.EXPECT().
					Publish(mock.Anything, []domainevent.Event{}).
					Return(nil)
			},
		},
		{
			name:   "propagates publisher error",
			events: fixedEvents,
			setupMock: func(ms *mockedService) {
				ms.publisher.EXPECT().
					Publish(mock.Anything, fixedEvents).
					Return(errors.New("kafka unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := newMockedService(t)
			tc.setupMock(ms)

			err := ms.Ingest(context.Background(), tc.events)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
