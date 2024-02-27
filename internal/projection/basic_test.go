package projection

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/zucchinho/ocpp/internal/domain"
	"github.com/zucchinho/ocpp/internal/domain/mock"
)

func TestNewBasicProjection(t *testing.T) {
	// arrange
	ctrl := gomock.NewController(t)
	mockEventSource := mock.NewMockEventSource(ctrl)

	// act
	bp := NewBasicProjection(mockEventSource)

	// assert
	assert.NotNil(t, bp)
	assert.Equal(t, mockEventSource, bp.eventSource)
}

func TestNumChargingStations(t *testing.T) {
	tests := []struct {
		name   string
		events []domain.Event
		want   int
	}{
		{
			name:   "no events",
			events: []domain.Event{},
			want:   0,
		},
		{
			name: "connector list request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    time.Now(),
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
				{
					ID:            "event-3",
					MessageID:     "message-3",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-2",
					},
				},
				{
					ID:            "event-4",
					MessageID:     "message-4",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    time.Now(),
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
			},
			want: 2,
		},
		{
			name: "meter values notification",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesNotification,
					OccurredAt:    time.Now(),
					Payload: domain.MeterValuesNotificationPayload{
						StationID: "station-1",
					},
				},
			},
			want: 1,
		},
		{
			name: "meter values request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    time.Now(),
					Payload: domain.MeterValuesResponsePayload{
						MeterValues: []domain.MeterValue{},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			mockEventSource := mock.NewMockEventSource(ctrl)
			bp := NewBasicProjection(mockEventSource)

			mockEventSource.EXPECT().GetAll(gomock.Any()).Return(tt.events)

			// act
			numChargingStations := bp.NumChargingStations(context.Background())

			// assert
			assert.Equal(t, tt.want, numChargingStations)
		})
	}

}

func TestNumConnectors(t *testing.T) {
	tests := []struct {
		name          string
		events        []domain.Event
		stationID     string
		correlationID string
		want          int
		wantErr       bool
	}{
		{
			name: "no events for station",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
			},
			stationID:     "station-2",
			correlationID: "",
			want:          0,
			wantErr:       true,
		},
		{
			name: "connector list request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    time.Now(),
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
			},
			stationID:     "station-1",
			correlationID: "correlation-1",
			want:          2,
		},
		{
			name: "meter values notification",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesNotification,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.MeterValuesNotificationPayload{
						StationID: "station-1",
						MeterValues: []domain.MeterValue{
							{
								ConnectorID: "connector-1",
								Reading:     "100",
							},
							{
								ConnectorID: "connector-2",
								Reading:     "200",
							},
							{
								ConnectorID: "connector-3",
								Reading:     "300",
							},
						},
					},
				},
			},
			stationID: "station-1",
			want:      3,
		},
		{
			name: "meter values request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesRequest,
					OccurredAt:    time.Now().Add(-time.Minute),
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    time.Now(),
					Payload: domain.MeterValuesResponsePayload{
						MeterValues: []domain.MeterValue{
							{
								ConnectorID: "connector-1",
								Reading:     "100",
							},
						},
					},
				},
			},
			stationID:     "station-1",
			correlationID: "correlation-1",
			want:          1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			mockEventSource := mock.NewMockEventSource(ctrl)
			bp := NewBasicProjection(mockEventSource)

			mockEventSource.EXPECT().GetAll(gomock.Any()).Return(tt.events)

			if tt.correlationID != "" {
				mockEventSource.EXPECT().GetByCorrelationID(gomock.Any(), tt.correlationID).Return(tt.events)
			}

			// act
			got, err := bp.NumConnectors(context.Background(), tt.stationID)

			// assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
