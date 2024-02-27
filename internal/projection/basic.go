package projection

import (
	"context"
	"errors"
	"time"

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

func (bp *BasicProjection) NumChargingStations(ctx context.Context) int {
	return len(bp.getStationIDs(ctx))
}

func (bp *BasicProjection) NumConnectors(ctx context.Context, stationID string) (int, error) {
	// Iterate through all events and find the latest event for the given stationID.
	latestEvents := bp.getLatestEventsForStationID(ctx, stationID)

	// If there are no events for the stationID, return an error.
	if len(latestEvents) == 0 {
		return 0, errors.New("charging station not found")
	}

	var latestRelevantEvent *domain.Event
	var numConnectors int
	for _, event := range latestEvents {
		// If the event is newer than the latest event, update the latest event.
		if latestRelevantEvent == nil || event.OccurredAt.After(latestRelevantEvent.OccurredAt) {
			switch event.MessageType {
			case domain.EventTypeConnectorListResponse:
				numConnectors = event.Payload.(domain.ConnectorListResponsePayload).NumConnectors
				latestRelevantEvent = &event
			case domain.EventTypeMeterValuesNotification:
				numConnectors = len(event.Payload.(domain.MeterValuesNotificationPayload).MeterValues)
				latestRelevantEvent = &event
			}
		}
	}

	return numConnectors, nil
}

func (bp *BasicProjection) ChargingStation(ctx context.Context, stationID string) (domain.ChargingStation, error) {
	latestEvents := bp.getLatestEventsForStationID(ctx, stationID)

	// If there are no events for the stationID, return an error.
	if len(latestEvents) == 0 {
		return domain.ChargingStation{}, errors.New("charging station not found")
	}

	// If there are events for the stationID, create a charging station from the latest events.
	var connectors []domain.Connector
	var latestEventTime time.Time

	// If there is a MeterValuesNotification event, use it to create the connectors.
	if meterValuesNotificationEvent, ok := latestEvents[domain.EventTypeMeterValuesNotification]; ok {
		payload := meterValuesNotificationEvent.Payload.(domain.MeterValuesNotificationPayload)
		for _, meterValue := range payload.MeterValues {
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
		payload := meterValuesResponseEvent.Payload.(domain.MeterValuesResponsePayload)
		for _, meterValue := range payload.MeterValues {
			var existingConnector *domain.Connector
			for _, connector := range connectors {
				if connector.ID == meterValue.ConnectorID {
					existingConnector = &connector
					break
				}
			}

			// If the connector exists and the MeterValuesResponse event is newer, update the connector.
			if existingConnector != nil && existingConnector.UpdatedAt.Before(meterValuesResponseEvent.OccurredAt) {
				existingConnector.Reading = meterValue.Reading
				existingConnector.UpdatedAt = meterValuesResponseEvent.OccurredAt
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
			numConnectors = connectorListResponseEvent.Payload.(domain.ConnectorListResponsePayload).NumConnectors

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

func (bp *BasicProjection) ChargingStations(ctx context.Context) []domain.ChargingStation {
	var chargingStations []domain.ChargingStation

	for _, stationID := range bp.getStationIDs(ctx) {
		chargingStation, err := bp.ChargingStation(ctx, stationID)
		if err == nil {
			chargingStations = append(chargingStations, chargingStation)
		}
	}

	return chargingStations
}

func (bp *BasicProjection) getStationIDs(ctx context.Context) []string {
	stationIDsMap := make(map[string]bool)
	stationIDs := make([]string, 0)

	for _, event := range bp.eventSource.GetAll(ctx) {
		var stationIDFromEvent string
		switch event.MessageType {
		case domain.EventTypeConnectorListRequest:
			stationIDFromEvent = event.Payload.(domain.ConnectorListRequestPayload).StationID
		case domain.EventTypeMeterValuesNotification:
			stationIDFromEvent = event.Payload.(domain.MeterValuesNotificationPayload).StationID
		case domain.EventTypeMeterValuesRequest:
			stationIDFromEvent = event.Payload.(domain.MeterValuesRequestPayload).StationID
		}

		// If the stationID is not in the map, add it.
		if !stationIDsMap[stationIDFromEvent] && stationIDFromEvent != "" {
			stationIDsMap[stationIDFromEvent] = true
			stationIDs = append(stationIDs, stationIDFromEvent)
		}
	}

	return stationIDs
}

func (bp *BasicProjection) getLatestEventsForStationID(ctx context.Context, stationID string) map[string]domain.Event {
	latestEvents := make(map[string]domain.Event, 0)

	// Iterate through all events and find the latest event for the given stationID.
	for _, event := range bp.eventSource.GetAll(ctx) {
		var eventIsForStationID bool
		switch event.MessageType {
		case domain.EventTypeConnectorListRequest:
			eventIsForStationID = event.Payload.(domain.ConnectorListRequestPayload).StationID == stationID
		case domain.EventTypeMeterValuesNotification:
			eventIsForStationID = event.Payload.(domain.MeterValuesNotificationPayload).StationID == stationID
		case domain.EventTypeMeterValuesRequest:
			eventIsForStationID = event.Payload.(domain.MeterValuesRequestPayload).StationID == stationID
		}

		if eventIsForStationID {
			if existingEventForType, ok := latestEvents[domain.EventTypeConnectorListRequest]; !ok || event.OccurredAt.After(existingEventForType.OccurredAt) {
				latestEvents[domain.EventTypeConnectorListRequest] = event
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

	return latestEvents
}
