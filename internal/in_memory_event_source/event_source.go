package inmemoryeventsource

import (
	"context"
	"fmt"
	"sync"

	"github.com/zucchinho/ocpp/internal/domain"
)

type InMemoryEventSource struct {
	mu     sync.Mutex
	events map[string]domain.Event
}

var _ domain.EventSource = &InMemoryEventSource{}

func NewInMemoryEventSource() *InMemoryEventSource {
	return &InMemoryEventSource{
		events: make(map[string]domain.Event),
	}
}

func (ies *InMemoryEventSource) Create(ctx context.Context, event domain.Event) (string, error) {
	ies.mu.Lock()

	if event.ID == "" {
		event.ID = "event-" + fmt.Sprint(len(ies.events))
	}

	ies.events[event.ID] = event

	ies.mu.Unlock()

	return event.ID, nil
}

func (ies *InMemoryEventSource) Get(ctx context.Context, id string) (domain.Event, error) {
	ies.mu.Lock()
	event, ok := ies.events[id]
	ies.mu.Unlock()

	if !ok {
		return domain.Event{}, domain.ErrEventNotFound
	}

	return event, nil
}

func (ies *InMemoryEventSource) GetByCorrelationID(ctx context.Context, correlationID string) ([]domain.Event, error) {
	ies.mu.Lock()
	defer ies.mu.Unlock()

	var events []domain.Event
	for _, event := range ies.events {
		if event.CorrelationID == correlationID {
			events = append(events, event)
		}
	}

	return events, nil
}
