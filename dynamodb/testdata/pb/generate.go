package pb

import (
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/savaki/snapsource"
)

//go:generate protoc --gogo_out=. events.proto

// AggregateID returns the id of the aggregate referenced by the event
func (m *UserCreated) AggregateID() string {
	return m.ID
}

// EventVersion contains the version number of this event
func (m *UserCreated) EventVersion() int {
	return int(m.Version)
}

// EventAt indicates when the event occurred
func (m *UserCreated) EventAt() time.Time {
	return time.Unix(m.At, 0)
}

type Serializer struct {
}

// MarshalEvent serializes an event
func (s Serializer) MarshalEvent(event snapsource.Event) ([]byte, error) {
	var payload Payload
	switch v := event.(type) {
	case *UserCreated:
		payload.T1 = v

	default:
		return nil, errors.Errorf("unhandled event type, %v", reflect.TypeOf(event))
	}

	data, err := proto.Marshal(&payload)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal %v", reflect.TypeOf(payload))
	}

	return data, nil
}

// UnmarshalEvent converts a []byte back into an Event
func (s Serializer) UnmarshalEvent(data []byte) (snapsource.Event, error) {
	var payload Payload
	if err := proto.Unmarshal(data, &payload); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal %v", reflect.TypeOf(payload))
	}

	switch payload.Type {
	case 1:
		return payload.T1, nil
	default:
		return nil, errors.Errorf("unhandled payload type")
	}
}
