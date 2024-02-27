package domain

import (
	"context"
	"time"
)

type ChargingStation struct {
	ID            string      `json:"id"`
	NumConnectors int         `json:"numConnectors"`
	Connectors    []Connector `json:"connectors"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

type Connector struct {
	ID                string    `json:"id"`
	ChargingStationID string    `json:"chargingStationId"`
	Reading           string    `json:"reading"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Store interface {
	UpsertConnector(ctx context.Context, connector Connector) (string, error)
	UpsertChargingStation(ctx context.Context, chargingStation ChargingStation) (string, error)
	GetChargingStation(ctx context.Context, id string) (ChargingStation, error)
	GetChargingStations(ctx context.Context) ([]ChargingStation, error)
}

type Projection interface {
	NumChargingStations(ctx context.Context) int
	NumConnectors(ctx context.Context, stationID string) (int, error)
	ChargingStation(ctx context.Context, stationID string) (ChargingStation, error)
	ChargingStations(ctx context.Context) []ChargingStation
}
