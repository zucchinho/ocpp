package domain

import "errors"

var (
	// ErrEventNotFound is returned when the event is not found.
	ErrEventNotFound = errors.New("event not found")
)
