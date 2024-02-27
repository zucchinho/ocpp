package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/zucchinho/ocpp/internal/domain"
)

type BasicProjection struct {
	eventSource domain.EventSource
}

var _ domain.Projection = &BasicProjection{}

func NewBasicProjection(eventSource domain.EventSource) *BasicProjection {
	return &BasicProjection{
		eventSource: eventSource,
	}
}

func (bp *BasicProjection) NumChargingStations(ctx context.Context) (int, error) {
	stationIDs, err := bp.getStationIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("get station IDs: %w", err)
	}
	return len(stationIDs), nil
}

func (bp *BasicProjection) NumConnectors(ctx context.Context, stationID string) (int, error) {
	// Iterate through all events and find the latest event for the given stationID.
	latestEvents, err := bp.getLatestEventsForStationID(ctx, stationID)
	if err != nil {
		return 0, fmt.Errorf("get latest events for station ID: %w", err)
	}

	// If there are no events for the stationID, return an error.
	if len(latestEvents) == 0 {
		return 0, errors.New("charging station not found")
	}

	var latestRelevantEvent *domain.Event
	var numConnectors int
	for _, event := range latestEvents {
		// If the event is newer than the latest event, update the latest event.
		if latestRelevantEvent == nil || event.OccurredAt.After(latestRelevantEvent.OccurredAt) {
			payload, err := convertEventPayload(event)
			if err != nil {
				return 0, fmt.Errorf("failed to convert event payload: %w", err)
			}
			switch event.MessageType {
			case domain.EventTypeConnectorListResponse:
				numConnectors = payload.(domain.ConnectorListResponsePayload).NumConnectors
				latestRelevantEvent = &event
			case domain.EventTypeMeterValuesNotification:
				numConnectors = len(payload.(domain.MeterValuesNotificationPayload).MeterValues)
				latestRelevantEvent = &event
			case domain.EventTypeMeterValuesResponse:
				numConnectors = len(payload.(domain.MeterValuesResponsePayload).MeterValues)
				latestRelevantEvent = &event
			}
		}
	}

	return numConnectors, nil
}

func (bp *BasicProjection) ChargingStation(ctx context.Context, stationID string) (domain.ChargingStation, error) {
	latestEvents, err := bp.getLatestEventsForStationID(ctx, stationID)
	if err != nil {
		return domain.ChargingStation{}, fmt.Errorf("get latest events for station ID: %w", err)
	}

	// If there are no events for the stationID, return an error.
	if len(latestEvents) == 0 {
		return domain.ChargingStation{}, errors.New("charging station not found")
	}

	// If there are events for the stationID, create a charging station from the latest events.
	var connectors []domain.Connector
	var latestEventTime time.Time

	// If there is a MeterValuesNotification event, use it to create the connectors.
	if meterValuesNotificationEvent, ok := latestEvents[domain.EventTypeMeterValuesNotification]; ok {
		payload, err := convertEventPayload(meterValuesNotificationEvent)
		if err != nil {
			return domain.ChargingStation{}, fmt.Errorf("failed to convert event payload: %w", err)
		}
		for _, meterValue := range payload.(domain.MeterValuesNotificationPayload).MeterValues {
			connectors = append(connectors, domain.Connector{
				ID:                meterValue.ConnectorID,
				ChargingStationID: stationID,
				Reading:           meterValue.Reading,
				UpdatedAt:         meterValuesNotificationEvent.OccurredAt,
			})
		}

		latestEventTime = meterValuesNotificationEvent.OccurredAt
	}

	// If there is a MeterValuesResponse event, use it to create/update the connectors.
	if meterValuesResponseEvent, ok := latestEvents[domain.EventTypeMeterValuesResponse]; ok {
		payload, err := convertEventPayload(meterValuesResponseEvent)
		if err != nil {
			return domain.ChargingStation{}, fmt.Errorf("failed to convert event payload: %w", err)
		}
		for _, meterValue := range payload.(domain.MeterValuesResponsePayload).MeterValues {
			var existingConnector *domain.Connector
			var connectorIdx int
			for i, connector := range connectors {
				if connector.ID == meterValue.ConnectorID {
					existingConnector = &connector
					connectorIdx = i
					break
				}
			}

			// If the connector exists and the MeterValuesResponse event is newer, update the connector.
			if existingConnector != nil && existingConnector.UpdatedAt.Before(meterValuesResponseEvent.OccurredAt) {
				existingConnector.Reading = meterValue.Reading
				existingConnector.UpdatedAt = meterValuesResponseEvent.OccurredAt

				connectors[connectorIdx] = *existingConnector
			} else {
				connectors = append(connectors, domain.Connector{
					ID:                meterValue.ConnectorID,
					ChargingStationID: stationID,
					Reading:           meterValue.Reading,
					UpdatedAt:         meterValuesResponseEvent.OccurredAt,
				})
			}
		}

		if meterValuesResponseEvent.OccurredAt.After(latestEventTime) {
			latestEventTime = meterValuesResponseEvent.OccurredAt
		}
	}

	// If there is a ConnectorListResponse event, use it to get the number of connectors.
	var numConnectors int
	if len(connectors) == 0 {
		if connectorListResponseEvent, ok := latestEvents[domain.EventTypeConnectorListResponse]; ok {
			payload, err := convertEventPayload(connectorListResponseEvent)
			if err != nil {
				return domain.ChargingStation{}, fmt.Errorf("failed to convert event payload: %w", err)
			}
			numConnectors = payload.(domain.ConnectorListResponsePayload).NumConnectors

			if connectorListResponseEvent.OccurredAt.After(latestEventTime) {
				latestEventTime = connectorListResponseEvent.OccurredAt
			}
		}
	} else {
		numConnectors = len(connectors)
	}

	return domain.ChargingStation{
		ID:            stationID,
		NumConnectors: numConnectors,
		Connectors:    connectors,
		UpdatedAt:     latestEventTime,
	}, nil
}

func (bp *BasicProjection) ChargingStations(ctx context.Context) ([]domain.ChargingStation, error) {
	var chargingStations []domain.ChargingStation

	stationIDs, err := bp.getStationIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get station IDs: %w", err)
	}

	for _, stationID := range stationIDs {
		chargingStation, err := bp.ChargingStation(ctx, stationID)
		if err == nil {
			chargingStations = append(chargingStations, chargingStation)
		}
	}

	return chargingStations, nil
}

func (bp *BasicProjection) getStationIDs(ctx context.Context) ([]string, error) {
	stationIDsMap := make(map[string]bool)
	stationIDs := make([]string, 0)

	for _, event := range bp.eventSource.GetAll(ctx) {
		payload, err := convertEventPayload(event)
		if err != nil {
			return nil, fmt.Errorf("convert event payload: %w", err)
		}

		var stationIDFromEvent string
		switch event.MessageType {
		case domain.EventTypeConnectorListRequest:
			stationIDFromEvent = payload.(domain.ConnectorListRequestPayload).StationID
		case domain.EventTypeMeterValuesNotification:
			stationIDFromEvent = payload.(domain.MeterValuesNotificationPayload).StationID
		case domain.EventTypeMeterValuesRequest:
			stationIDFromEvent = payload.(domain.MeterValuesRequestPayload).StationID
		}

		// If the stationID is not in the map, add it.
		if !stationIDsMap[stationIDFromEvent] && stationIDFromEvent != "" {
			stationIDsMap[stationIDFromEvent] = true
			stationIDs = append(stationIDs, stationIDFromEvent)
		}
	}

	return stationIDs, nil
}

func (bp *BasicProjection) getLatestEventsForStationID(ctx context.Context, stationID string) (map[string]domain.Event, error) {
	latestEvents := make(map[string]domain.Event, 0)

	// Iterate through all events and find the latest event for the given stationID.
	for _, event := range bp.eventSource.GetAll(ctx) {
		payload, err := convertEventPayload(event)
		if err != nil {
			return nil, fmt.Errorf("convert event payload: %w", err)
		}

		var eventIsForStationID bool
		switch event.MessageType {
		case domain.EventTypeConnectorListRequest:
			eventIsForStationID = payload.(domain.ConnectorListRequestPayload).StationID == stationID
		case domain.EventTypeMeterValuesNotification:
			eventIsForStationID = payload.(domain.MeterValuesNotificationPayload).StationID == stationID
		case domain.EventTypeMeterValuesRequest:
			eventIsForStationID = payload.(domain.MeterValuesRequestPayload).StationID == stationID
		}

		if eventIsForStationID {
			if existingEventForType, ok := latestEvents[domain.EventTypeConnectorListRequest]; !ok || event.OccurredAt.After(existingEventForType.OccurredAt) {
				latestEvents[event.MessageType] = event
			}
		}
	}

	if len(latestEvents) > 0 {
		// If there is a connector list request event, find the latest corresponding response event (with the same correlation ID).
		if connectorListRequestEvent, ok := latestEvents[domain.EventTypeConnectorListRequest]; ok {
			// Find the latest corresponding response event (with the same correlation ID).
			var latestCorrespondingResponseEvent *domain.Event
			for _, event := range bp.eventSource.GetByCorrelationID(ctx, connectorListRequestEvent.CorrelationID) {
				if event.MessageType == domain.EventTypeConnectorListResponse && (latestCorrespondingResponseEvent == nil || event.OccurredAt.After(latestCorrespondingResponseEvent.OccurredAt)) {
					latestCorrespondingResponseEvent = &event
				}
			}

			if latestCorrespondingResponseEvent != nil {
				latestEvents[domain.EventTypeConnectorListResponse] = *latestCorrespondingResponseEvent
			}
		}

		// If there is a meter values request event, find the latest corresponding response event (with the same correlation ID).
		if meterValuesRequestEvent, ok := latestEvents[domain.EventTypeMeterValuesRequest]; ok {
			// Find the latest corresponding response event (with the same correlation ID).
			var latestCorrespondingResponseEvent *domain.Event
			for _, event := range bp.eventSource.GetByCorrelationID(ctx, meterValuesRequestEvent.CorrelationID) {
				if event.MessageType == domain.EventTypeMeterValuesResponse && (latestCorrespondingResponseEvent == nil || event.OccurredAt.After(latestCorrespondingResponseEvent.OccurredAt)) {
					latestCorrespondingResponseEvent = &event
				}
			}

			if latestCorrespondingResponseEvent != nil {
				latestEvents[domain.EventTypeMeterValuesResponse] = *latestCorrespondingResponseEvent
			}
		}
	}

	return latestEvents, nil
}

func convertEventPayload(event domain.Event) (any, error) {
	switch event.MessageType {
	case domain.EventTypeMeterValuesRequest:
		payload := domain.MeterValuesRequestPayload{}
		if err := mapstructure.Decode(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	case domain.EventTypeMeterValuesResponse:
		payload := domain.MeterValuesResponsePayload{}
		if err := mapstructure.Decode(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	case domain.EventTypeMeterValuesNotification:
		payload := domain.MeterValuesNotificationPayload{}
		if err := mapstructure.Decode(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	case domain.EventTypeConnectorListRequest:
		payload := domain.ConnectorListRequestPayload{}
		if err := mapstructure.Decode(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	case domain.EventTypeConnectorListResponse:
		payload := domain.ConnectorListResponsePayload{}
		if err := mapstructure.Decode(event.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", event.MessageType)
	}
}
