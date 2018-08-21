package esdb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/tj/assert"
)

type User struct {
}

func (u *User) Apply(ctx context.Context, command Command) ([]Event, error) {
	fmt.Printf("applied %#v\n", command)
	return nil, nil
}

type Mock struct {
}

func (m Mock) AggregateID() string {
	return "abc"
}

func TestHandler(t *testing.T) {
	withTable(t, func(api *dynamodb.DynamoDB, tableName string) {
		h, err := New(Config{
			API:       api,
			TableName: tableName,
			Prototype: reflect.TypeOf(&User{}),
		})
		assert.Nil(t, err)

		h.Apply(context.Background(), Mock{})
	})
}
