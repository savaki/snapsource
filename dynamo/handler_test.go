package dynamo

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/neelance/graphql-go/errors"
	"github.com/savaki/snapsource"
	"github.com/savaki/snapsource/dynamo/testdata/pb"
	"github.com/tj/assert"
)

type User struct {
	Version int
	Name    string
	Email   string
}

func (u *User) On(event snapsource.Event) error {
	switch v := event.(type) {
	case *pb.UserCreated:
		u.Email = v.Email
		u.Name = v.Name

	case *pb.EmailUpdated:
		u.Email = v.NewEmail

	default:
		return errors.Errorf("unable to handle event, %v", reflect.TypeOf(event))
	}

	u.Version = event.EventVersion()

	return nil
}

func (u *User) Apply(ctx context.Context, command snapsource.Command) ([]snapsource.Event, error) {
	b := pb.NewBuilder(command.AggregateID(), u.Version)

	switch cmd := command.(type) {
	case *pb.CreateUser:
		b.UserCreated(cmd.Name, cmd.Email)

	case *pb.UpdateEmail:
		b.EmailUpdated(u.Email, cmd.NewEmail)

	default:
		return nil, errors.Errorf("unable to apply command")
	}

	return b.Events, nil
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

		ctx := context.Background()
		id := "abc"

		events, err := h.Apply(ctx, &pb.CreateUser{
			ID:    id,
			Name:  "Joe Public",
			Email: "job.public@example.com",
		})
		assert.Nil(t, err)
		assert.Len(t, events, 1)

		events, err = h.Apply(ctx, &pb.UpdateEmail{
			ID:       id,
			NewEmail: "joe.public@example.com",
		})
		assert.Nil(t, err)
		assert.Len(t, events, 1)

		var user User
		meta, err := h.Load(ctx, id, &user)
		assert.Nil(t, err)
		assert.Equal(t, id, meta.ID)
		assert.Equal(t, 2, meta.Version)
		assert.True(t, time.Now().Sub(meta.CreatedAt) < 3*time.Second)
		assert.True(t, time.Now().Sub(meta.UpdatedAt) < 3*time.Second)

		assert.Equal(t, 2, user.Version)
		assert.Equal(t, "Joe Public", user.Name)
		assert.Equal(t, "joe.public@example.com", user.Email)
	})
}
