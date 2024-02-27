package processor

import (
	"context"

	"github.com/zucchinho/ocpp/internal/domain"
)

type eventProcessor struct {
	eventSource domain.EventSource
}

func NewEventProcessor(
	eventSource domain.EventSource,
) domain.EventProcessor {
	return &eventProcessor{
		eventSource: eventSource,
	}
}

func (ep *eventProcessor) ProcessEvent(ctx context.Context, event domain.Event) error {
	ep.eventSource.Create(ctx, event)

	return nil
}
