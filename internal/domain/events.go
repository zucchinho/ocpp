package domain

import (
	"context"
	"time"
)

const (
	EventTypeMeterValuesRequest      = "MeterValuesRequest"
	EventTypeMeterValuesResponse     = "MeterValuesResponse"
	EventTypeMeterValuesNotification = "MeterValuesNotification"
	EventTypeConnectorListRequest    = "ConnectorListRequest"
	EventTypeConnectorListResponse   = "ConnectorListResponse"
)

// MeterValuesRequestPayload is the payload for the MeterValuesRequest event.
type MeterValuesRequestPayload struct {
	StationID   string `json:"stationId"`
	ConnectorID string `json:"connectorId"`
}

// MeterValue is a single meter value containing the reading for a connector.
type MeterValue struct {
	ConnectorID string `json:"connectorId"`
	Reading     string `json:"reading"`
}

// MeterValuesRequestPayload is the payload for the MeterValuesRequest event.
type MeterValuesResponsePayload struct {
	MeterValues []MeterValue `json:"meterValues"`
}

// MeterValuesNotificationPayload is the payload for the MeterValuesNotification event.
type MeterValuesNotificationPayload struct {
	StationID   string       `json:"stationId"`
	MeterValues []MeterValue `json:"meterValues"`
}

// ConnectorListRequestPayload is the payload for the ConnectorListRequest event.
type ConnectorListRequestPayload struct {
	StationID string `json:"stationId"`
}

// ConnectorListResponsePayload is the payload for the ConnectorListResponse event.
type ConnectorListResponsePayload struct {
	NumConnectors int `json:"numConnectors"`
}

type Event struct {
	ID            string    `json:"id"`
	MessageID     string    `json:"messageId"`
	CorrelationID string    `json:"correlationId"`
	MessageType   string    `json:"messageType"`
	OccurredAt    time.Time `json:"occurredAt"`
	Payload       any       `json:"payload"`
}

type EventProcessor interface {
	// ProcessEvent processes the given event.
	ProcessEvent(ctx context.Context, event Event) error
}

// Source of events, each event is unique and has a unique ID.
type EventSource interface {
	// Get returns the event by the ID.
	Get(ctx context.Context, correlationID string) (Event, error)
	// Create creates a new event, returning the event's unique ID.
	Create(ctx context.Context, event Event) (string, error)
	// GetByCorrelationID returns all events with the given correlation ID.
	GetByCorrelationID(ctx context.Context, correlationID string) []Event
	// GetAll returns all events.
	GetAll(ctx context.Context) []Event
}
