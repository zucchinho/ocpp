# Event Processor

Creates events in the underlying event source

# Event Source

## In Memory

A placeholder implementation which simply stores events in memory for them to be accessed by the same program.

# Projection

## Basic

A basic projection which calculates the current state of the charging stations based on the events currently in the source

# Building

```sh
go build cmd/cli/json_event_consumer/main.go
```

# Running

```sh
./main -input events.json
```
