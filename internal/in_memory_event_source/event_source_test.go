package inmemoryeventsource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zucchinho/ocpp/internal/domain"
)

func TestInMemoryEventSource_Create_NoID(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()

	// act
	id, err := ies.Create(context.Background(), domain.Event{
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "event-1", id)
}

func TestInMemoryEventSource_Create_WithID(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()

	// act
	id, err := ies.Create(context.Background(), domain.Event{
		ID:            "event-1",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "event-1", id)
}

func TestInMemoryEventSource_Get(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()
	id, _ := ies.Create(context.Background(), domain.Event{
		ID:            "event-1",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})

	// act
	event, err := ies.Get(context.Background(), id)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, domain.Event{
		ID:            "event-1",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	}, event)
}

func TestInMemoryEventSource_Get_NotFound(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()

	// act
	event, err := ies.Get(context.Background(), "event-1")

	// assert
	assert.Equal(t, domain.Event{}, event)
	assert.Equal(t, domain.ErrEventNotFound, err)
}

func TestInMemoryEventSource_GetByCorrelationID(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()
	ies.Create(context.Background(), domain.Event{
		ID:            "event-1",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})
	ies.Create(context.Background(), domain.Event{
		ID:            "event-2",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})

	// act
	events := ies.GetByCorrelationID(context.Background(), "12345")

	// assert
	assert.ElementsMatch(t, []domain.Event{
		{
			ID:            "event-1",
			MessageID:     "1",
			MessageType:   "MessageType",
			CorrelationID: "12345",
			Payload:       []byte("payload"),
		},
		{
			ID:            "event-2",
			MessageID:     "1",
			MessageType:   "MessageType",
			CorrelationID: "12345",
			Payload:       []byte("payload"),
		},
	}, events)
}

func TestInMemoryEventSource_GetAll(t *testing.T) {
	// arrange
	ies := NewInMemoryEventSource()
	ies.Create(context.Background(), domain.Event{
		ID:            "event-1",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})
	ies.Create(context.Background(), domain.Event{
		ID:            "event-2",
		MessageID:     "1",
		MessageType:   "MessageType",
		CorrelationID: "12345",
		Payload:       []byte("payload"),
	})

	// act
	events := ies.GetAll(context.Background())

	// assert
	assert.ElementsMatch(t, []domain.Event{
		{
			ID:            "event-1",
			MessageID:     "1",
			MessageType:   "MessageType",
			CorrelationID: "12345",
			Payload:       []byte("payload"),
		},
		{
			ID:            "event-2",
			MessageID:     "1",
			MessageType:   "MessageType",
			CorrelationID: "12345",
			Payload:       []byte("payload"),
		},
	}, events)
}
