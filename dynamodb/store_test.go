package dynamodb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
	"github.com/savaki/snapsource"
	"github.com/tj/assert"
)

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(errNotFound))
	assert.True(t, IsNotFound(errors.Wrap(errNotFound, "1")))
	assert.True(t, IsNotFound(errors.Wrap(errors.Wrap(errNotFound, "1"), "2")))
	assert.False(t, IsNotFound(io.EOF))
	assert.False(t, IsNotFound(nil))
}

func withTable(t *testing.T, callback func(api *dynamodb.DynamoDB, tableName string)) {
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("blah", "blah", "blah"),
		Region:      aws.String("us-west-2"),
		Endpoint:    aws.String("http://localhost:8000"),
	})
	assert.Nil(t, err)

	api := dynamodb.New(s)

	content := make([]byte, 8)
	rand.Read(content)
	tableName := "tmp-" + hex.EncodeToString(content)
	input := dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(fieldID),
				AttributeType: aws.String(dynamodb.ScalarAttributeTypeS),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(fieldID),
				KeyType:       aws.String(dynamodb.KeyTypeHash),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(3),
			WriteCapacityUnits: aws.Int64(3),
		},
	}

	_, err = api.CreateTable(&input)
	assert.Nil(t, err)

	err = api.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	assert.Nil(t, err)

	defer api.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})

	callback(api, tableName)
}

func TestLifecycle(t *testing.T) {
	ctx := context.Background()

	t.Run("create then update", func(t *testing.T) {
		withTable(t, func(api *dynamodb.DynamoDB, tableName string) {
			d := Store{
				api:       api,
				tableName: tableName,
			}

			// create ----------------------

			want := createIn{
				ID:      "abc",
				Payload: &dynamodb.AttributeValue{S: aws.String("hello world")},
			}
			err := d.create(ctx, want)
			assert.Nil(t, err)

			err = d.create(ctx, want)
			assert.NotNil(t, err) // duplicates should be prevented

			got, ok, err := d.get(ctx, want.ID)
			assert.Nil(t, err)
			assert.True(t, ok)
			assert.Equal(t, 1, got.Version)
			assert.Equal(t, want.ID, got.ID)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.True(t, time.Now().Sub(got.CreatedAt) <= 3*time.Second)
			assert.True(t, time.Now().Sub(got.UpdatedAt) <= 3*time.Second)

			// update ----------------------

			err = d.update(ctx, updateIn{
				ID:      want.ID,
				Version: got.Version,
				Payload: &dynamodb.AttributeValue{S: aws.String("argle bargle")},
			})
			assert.Nil(t, err)

			got, ok, err = d.get(ctx, want.ID)
			assert.Nil(t, err)
			assert.True(t, ok)
		})
	})
}

func TestImplementsCommandHandler(t *testing.T) {
	var h snapsource.CommandHandler = &Handler{}
	assert.NotNil(t, h)
}
