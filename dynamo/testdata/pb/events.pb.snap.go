// Code generated by snapsource. DO NOT EDIT.
// source: events.proto

package pb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/savaki/snapsource"
	"github.com/gogo/protobuf/proto"
)

type Serializer struct {
}

func (s Serializer) MarshalEvent(event snapsource.Event) ([]byte, error) {
	return MarshalEvent(event)
}

func (s Serializer) UnmarshalEvent(data []byte) (snapsource.Event, error) {
	return UnmarshalEvent(data)
}

func NewSerializer() snapsource.Serializer {
	return Serializer{}
}

func (m *UserCreated) AggregateID() string { return m.ID }
func (m *UserCreated) EventVersion() int32 { return m.Version }
func (m *UserCreated) EventAt() time.Time  { return time.Unix(m.At, 0) }

func (m *EmailUpdated) AggregateID() string { return m.ID }
func (m *EmailUpdated) EventVersion() int32 { return m.Version }
func (m *EmailUpdated) EventAt() time.Time  { return time.Unix(m.At, 0) }


func MarshalEvent(event snapsource.Event) ([]byte, error) {
	payload := &Payload{}

	switch v := event.(type) {

	case *UserCreated:
		payload.Type = 2
		payload.T1 = v

	case *EmailUpdated:
		payload.Type = 3
		payload.T2 = v

	default:
		return nil, fmt.Errorf("Unhandled type, %v", event)
	}

	data, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func UnmarshalEvent(data []byte) (snapsource.Event, error) {
	container := &Payload{};
	err := proto.Unmarshal(data, container)
	if err != nil {
		return nil, err
	}

	var event interface{}
	switch container.Type {

	case 2:
		event = container.T1

	case 3:
		event = container.T2

	default:
		return nil, fmt.Errorf("Unhandled type, %v", container.Type)
	}

	return event.(snapsource.Event), nil
}

type Encoder struct{
	w io.Writer
}

func (e *Encoder) WriteEvent(event snapsource.Event) (int, error) {
	data, err := MarshalEvent(event)
	if err != nil {
		return 0, err
	}

	// Write the length of the marshaled event as uint64
	//
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(len(data)))
	if _, err := e.w.Write(buffer); err != nil {
		return 0, err
	}

	n, err := e.w.Write(data)
	if err != nil {
		return 0, err
	}

	return n + 8, nil
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

type Decoder struct {
	r       *bufio.Reader
	scratch *bytes.Buffer
}

func (d *Decoder) readN(n uint64) ([]byte, error) {
	d.scratch.Reset()
	for i := uint64(0); i < n; i++ {
		b, err := d.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if err := d.scratch.WriteByte(b); err != nil {
			return nil, err
		}
	}
	return d.scratch.Bytes(), nil
}

func (d *Decoder) ReadEvent() (snapsource.Event, error) {
	data, err := d.readN(8)
	if err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint64(data)

	data, err = d.readN(length)
	if err != nil {
		return nil, err
	}

	event, err := UnmarshalEvent(data)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder {
		r:       bufio.NewReader(r),
		scratch: bytes.NewBuffer(nil),
	}
}

type Builder struct {
	id      string
	version int32
	Events  []snapsource.Event
}

func NewBuilder(id string, version int32) *Builder {
	return &Builder {
		id:      id,
		version: version,
	}
}

func (b *Builder) nextVersion() int32 {
	b.version++
	return b.version
}


func (b *Builder) UserCreated(name string, email string, ) {
	event := &UserCreated{
		ID:      b.id,
		Version: b.nextVersion(),
		At:      time.Now().Unix(),
	Name: name,
	Email: email,

	}
	b.Events = append(b.Events, event)
}

func (b *Builder) EmailUpdated(oldEmail string, newEmail string, ) {
	event := &EmailUpdated{
		ID:      b.id,
		Version: b.nextVersion(),
		At:      time.Now().Unix(),
	OldEmail: oldEmail,
	NewEmail: newEmail,

	}
	b.Events = append(b.Events, event)
}
