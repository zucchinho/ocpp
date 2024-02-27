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

func TestNumConnectors(t *testing.T) {
	// arrange
	ctrl := gomock.NewController(t)
	mockEventSource := mock.NewMockEventSource(ctrl)
	bp := NewBasicProjection(mockEventSource)

	events := []domain.Event{
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
	}
	mockEventSource.EXPECT().GetAll(gomock.Any()).Return(events)
	mockEventSource.EXPECT().GetByCorrelationID(gomock.Any(), "correlation-1").Return(events)

	// act
	numConnectors, err := bp.NumConnectors(context.Background(), "station-1")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, 2, numConnectors)
}
