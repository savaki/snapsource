package esdb

import (
	"context"
	"time"
)

// Event describe a change that happened to the Aggregate
//
// * Past tense e.g. EmailChanged
// * Contains intent e.g. EmailChanged is better than EmailSet
type Event interface {
	// AggregateID returns the id of the aggregate referenced by the event
	AggregateID() string

	// EventVersion contains the version number of this event
	EventVersion() int

	// EventAt indicates when the event occurred
	EventAt() time.Time
}

// Command encapsulates the data to mutate an aggregate
type Command interface {
	// AggregateID represents the id of the aggregate to apply to
	AggregateID() string
}

// Aggregate represents the aggregate root in the domain driven design sense.
// It represents the current state of the domain object and can be thought of
// as a left fold over events.
type Aggregate interface {
	// On will be called for each event; returns err if the event could not be
	// applied
	On(event Event) error
}

// CommandHandler consumes a command and emits Events
type CommandHandler interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Apply(ctx context.Context, command Command) ([]Event, error)
}

// Serializer converts between Events and Records
type Serializer interface {
	// MarshalEvent serializes an event
	MarshalEvent(event Event) ([]byte, error)

	// UnmarshalEvent converts a []byte back into an Event
	UnmarshalEvent(data []byte) (Event, error)
}
