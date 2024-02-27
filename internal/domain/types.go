package domain

import (
	"context"
	"time"
)

const (
	EventTypeMeterValuesRequest  = "MeterValuesRequest"
	EventTypeMeterValuesResponse = "MeterValuesResponse"
)

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

type ChargingStation struct {
	ID            string      `json:"id"`
	NumConnectors int         `json:"numConnectors"`
	Connectors    []Connector `json:"connectors"`
}

type Connector struct {
	ID                string `json:"id"`
	ChargingStationID string `json:"chargingStationId"`
	Reading           string `json:"reading"`
}

// Source of events, each event is unique and has a unique ID.
type EventSource interface {
	// Get returns the event by the ID.
	Get(ctx context.Context, correlationID string) (Event, error)
	// Create creates a new event, returning the event's unique ID.
	Create(ctx context.Context, event Event) (string, error)
	// GetByCorrelationID returns all events with the given correlation ID.
	GetByCorrelationID(ctx context.Context, correlationID string) ([]Event, error)
}

type Store interface {
	UpsertConnector(ctx context.Context, connector Connector) (string, error)
	UpsertChargingStation(ctx context.Context, chargingStation ChargingStation) (string, error)
	GetChargingStation(ctx context.Context, id string) (ChargingStation, error)
	GetChargingStations(ctx context.Context) ([]ChargingStation, error)
}
