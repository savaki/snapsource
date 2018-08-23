package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/savaki/snapsource"
	"github.com/savaki/snapsource/dynamodb/testdata/pb"
	"github.com/tj/assert"
)

type User struct {
	Email string
}

func (u *User) On(event snapsource.Event) error {
	return nil
}

func (u *User) Apply(ctx context.Context, command snapsource.Command) ([]snapsource.Event, error) {
	return []snapsource.Event{
		&pb.UserCreated{
			ID:      "a",
			Version: 1,
			At:      2,
		},
	}, nil
}

type Mock struct {
}

func (m Mock) AggregateID() string {
	return "abc"
}

func TestHandler(t *testing.T) {
	withTable(t, func(api *dynamodb.DynamoDB, tableName string) {
		h, err := New(Config{
			API:        api,
			TableName:  tableName,
			Serializer: pb.Serializer{},
			Factory:    func(id string) snapsource.Prototype { return &User{} },
		})
		assert.Nil(t, err)

		events, err := h.Apply(context.Background(), Mock{})
		assert.Nil(t, err)
		assert.Len(t, events, 1)
	})
}
