package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"github.com/zucchinho/ocpp/internal/domain"
	processor "github.com/zucchinho/ocpp/internal/event_processor"
	inmemoryeventsource "github.com/zucchinho/ocpp/internal/in_memory_event_source"
	"github.com/zucchinho/ocpp/internal/projection"
)

func main() {
	ctx := context.Background()
	var inputFlag = flag.String("input", "", "input file")
	flag.Parse()

	if *inputFlag == "" {
		flag.PrintDefaults()
		return
	}

	input := *inputFlag

	jsonFile, err := os.Open(input)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer jsonFile.Close()
	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var events []domain.Event

	if err := json.Unmarshal(jsonBytes, &events); err != nil {
		log.Fatalf("failed to unmarshal json: %v", err)
	}

	eventSource := inmemoryeventsource.NewInMemoryEventSource()
	eventProcessor := processor.NewEventProcessor(
		eventSource,
	)

	var errEventProcessing error
	for _, event := range events {
		if err := eventProcessor.ProcessEvent(ctx, event); err != nil {
			log.Printf("failed to process event: %v", err)
			errEventProcessing = errors.Join(errEventProcessing, err)
		}
	}

	if errEventProcessing != nil {
		log.Fatalf("failed to process events: %v", errEventProcessing)
	}

	log.Printf("processed %d events\n", len(events))

	views := projection.NewBasicProjection(eventSource)

	// print the number of charging stations
	numChargingStations, err := views.NumChargingStations(ctx)
	if err != nil {
		log.Fatalf("failed to get number of charging stations: %v", err)
	}
	log.Printf("num charging stations: %d\n", numChargingStations)

	chargingStations, err := views.ChargingStations(ctx)
	if err != nil {
		log.Fatalf("failed to get charging stations: %v", err)
	}

	for _, station := range chargingStations {
		stationJSON, err := json.MarshalIndent(station, "", "  ")
		if err != nil {
			log.Fatalf("failed to marshal station: %v", err)
		}
		log.Printf("charging station: %s\n", stationJSON)
	}
}
