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

var now = time.Now().Round(0)
var oneMinuteAgo = now.Add(-1 * time.Minute)
var twoMinutesAgo = now.Add(-2 * time.Minute)

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
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
				{
					ID:            "event-3",
					MessageID:     "message-3",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-2",
					},
				},
				{
					ID:            "event-4",
					MessageID:     "message-4",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
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
					OccurredAt:    now,
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
					OccurredAt:    oneMinuteAgo,
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    now,
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
					OccurredAt:    oneMinuteAgo,
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
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
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
					OccurredAt:    oneMinuteAgo,
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
					OccurredAt:    oneMinuteAgo,
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    now,
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
				mockEventSource.EXPECT().GetByCorrelationID(gomock.Any(), tt.correlationID).
					Return(getCorrelatedEvents(tt.events, tt.correlationID))
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

func TestChargingStation(t *testing.T) {
	tests := []struct {
		name          string
		events        []domain.Event
		stationID     string
		correlationID string
		want          domain.ChargingStation
		wantErr       bool
	}{
		{
			name: "no events for station",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    now,
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
			stationID:     "station-2",
			correlationID: "",
			want:          domain.ChargingStation{},
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
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
			},
			stationID:     "station-1",
			correlationID: "correlation-1",
			want: domain.ChargingStation{
				ID:            "station-1",
				NumConnectors: 2,
				UpdatedAt:     now,
			},
		},
		{
			name: "meter values notification",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesNotification,
					OccurredAt:    oneMinuteAgo,
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
			stationID:     "station-1",
			correlationID: "",
			want: domain.ChargingStation{
				ID:            "station-1",
				NumConnectors: 3,
				UpdatedAt:     oneMinuteAgo,
				Connectors: []domain.Connector{
					{
						ID:                "connector-1",
						ChargingStationID: "station-1",
						Reading:           "100",
						UpdatedAt:         oneMinuteAgo,
					},
					{
						ID:                "connector-2",
						ChargingStationID: "station-1",
						Reading:           "200",
						UpdatedAt:         oneMinuteAgo,
					},
					{
						ID:                "connector-3",
						ChargingStationID: "station-1",
						Reading:           "300",
						UpdatedAt:         oneMinuteAgo,
					},
				},
			},
		},
		{
			name: "meter values request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    now,
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
			want: domain.ChargingStation{
				ID:            "station-1",
				NumConnectors: 1,
				UpdatedAt:     now,
				Connectors: []domain.Connector{
					{
						ID:                "connector-1",
						ChargingStationID: "station-1",
						Reading:           "100",
						UpdatedAt:         now,
					},
				},
			},
		},
		{
			name: "meter values notification and meter values request/response: connector reading updated",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeMeterValuesNotification,
					OccurredAt:    twoMinutesAgo,
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
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeMeterValuesRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.MeterValuesRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeMeterValuesResponse,
					OccurredAt:    now,
					Payload: domain.MeterValuesResponsePayload{
						MeterValues: []domain.MeterValue{
							{
								ConnectorID: "connector-1",
								Reading:     "120",
							},
						},
					},
				},
			},
			stationID:     "station-1",
			correlationID: "correlation-2",
			want: domain.ChargingStation{
				ID:            "station-1",
				NumConnectors: 3,
				UpdatedAt:     now,
				Connectors: []domain.Connector{
					{
						ID:                "connector-1",
						ChargingStationID: "station-1",
						// The reading from the later MeterValuesResponse event should be used.
						Reading:   "120",
						UpdatedAt: now,
					},
					{
						ID:                "connector-2",
						ChargingStationID: "station-1",
						Reading:           "200",
						UpdatedAt:         twoMinutesAgo,
					},
					{
						ID:                "connector-3",
						ChargingStationID: "station-1",
						Reading:           "300",
						UpdatedAt:         twoMinutesAgo,
					},
				},
			},
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
				mockEventSource.EXPECT().GetByCorrelationID(gomock.Any(), tt.correlationID).
					Return(getCorrelatedEvents(tt.events, tt.correlationID))
			}

			// act
			got, err := bp.ChargingStation(context.Background(), tt.stationID)

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

func TestChargingStations(t *testing.T) {
	tests := []struct {
		name           string
		events         []domain.Event
		correlationIDs []string
		want           []domain.ChargingStation
	}{
		{
			name:   "no events",
			events: []domain.Event{},
			want:   nil,
		},
		{
			name: "connector list request/response",
			events: []domain.Event{
				{
					ID:            "event-1",
					MessageID:     "message-1",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-1",
					},
				},
				{
					ID:            "event-2",
					MessageID:     "message-2",
					CorrelationID: "correlation-1",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
				{
					ID:            "event-3",
					MessageID:     "message-3",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListRequest,
					OccurredAt:    oneMinuteAgo,
					Payload: domain.ConnectorListRequestPayload{
						StationID: "station-2",
					},
				},
				{
					ID:            "event-4",
					MessageID:     "message-4",
					CorrelationID: "correlation-2",
					MessageType:   domain.EventTypeConnectorListResponse,
					OccurredAt:    now,
					Payload: domain.ConnectorListResponsePayload{
						NumConnectors: 2,
					},
				},
			},
			correlationIDs: []string{"correlation-1", "correlation-2"},
			want: []domain.ChargingStation{
				{
					ID:            "station-1",
					NumConnectors: 2,
					UpdatedAt:     now,
				},
				{
					ID:            "station-2",
					NumConnectors: 2,
					UpdatedAt:     now,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			ctrl := gomock.NewController(t)
			mockEventSource := mock.NewMockEventSource(ctrl)
			bp := NewBasicProjection(mockEventSource)

			mockEventSource.EXPECT().GetAll(gomock.Any()).Return(tt.events).
				Times(len(tt.want) + 1) // +1 for the getStationIDs call
			for _, correlationID := range tt.correlationIDs {
				mockEventSource.EXPECT().GetByCorrelationID(gomock.Any(), correlationID).
					Return(getCorrelatedEvents(tt.events, correlationID))
			}

			// act
			got := bp.ChargingStations(context.Background())

			// assert
			assert.Equal(t, tt.want, got)
		})
	}
}

func getCorrelatedEvents(events []domain.Event, correlationID string) []domain.Event {
	correlatedEvents := make([]domain.Event, 0)
	for _, event := range events {
		if event.CorrelationID == correlationID {
			correlatedEvents = append(correlatedEvents, event)
		}
	}
	return correlatedEvents
}
