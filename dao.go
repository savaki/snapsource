package snapsource

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

const (
	fieldID        = "id"
	fieldVersion   = "version"
	fieldCreatedAt = "createdAt"
	fieldUpdatedAt = "updatedAt"
	fieldEvents    = "events"
	fieldPayload   = "payload"
)

const (
	dateLayout = time.RFC3339
)

type state struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Payload   *dynamodb.AttributeValue
}

func (s *state) Unmarshal(item map[string]*dynamodb.AttributeValue) error {
	if item == nil {
		return nil
	}

	id, err := unmarshalString(item[fieldID])
	if err != nil {
		return err
	}

	version, err := unmarshalInt(item[fieldVersion])
	if err != nil {
		return err
	}

	createdAt, err := unmarshalTime(item[fieldCreatedAt])
	if err != nil {
		return err
	}

	updatedAt, err := unmarshalTime(item[fieldUpdatedAt])
	if err != nil {
		return err
	}

	value := state{
		ID:        id,
		Version:   version,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Payload:   item[fieldPayload],
	}

	*s = value
	return nil
}

type dao struct {
	api        dynamodbiface.DynamoDBAPI
	tableName  string
	serializer Serializer
}

func (d dao) Get(ctx context.Context, id string) (state, bool, error) {
	input := dynamodb.GetItemInput{
		TableName:      aws.String(d.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			fieldID: {S: aws.String(id)},
		},
	}
	out, err := d.api.GetItemWithContext(ctx, &input)
	if err != nil {
		return state{}, false, errors.Wrapf(err, "unable to retrieve item from %v, id: %v", d.tableName)
	}
	if len(out.Item) == 0 {
		return state{}, false, nil
	}

	var v state
	if err := v.Unmarshal(out.Item); err != nil {
		return state{}, false, err
	}

	return v, true, nil
}

func (d dao) marshalEvents(events []Event) (*dynamodb.AttributeValue, error) {
	var datum [][]byte
	for _, event := range events {
		data, err := d.serializer.MarshalEvent(event)
		if err != nil {
			return nil, err
		}
		datum = append(datum, data)
	}

	if len(datum) == 0 {
		return nil, nil
	}

	return &dynamodb.AttributeValue{BS: datum}, nil
}

type createIn struct {
	ID      string
	Payload *dynamodb.AttributeValue
	Events  []Event
}

func (d dao) create(ctx context.Context, in createIn) error {
	events, err := d.marshalEvents(in.Events)
	if err != nil {
		return err
	}

	now := time.Now().Format(dateLayout)
	item := map[string]*dynamodb.AttributeValue{
		fieldID:        {S: aws.String(in.ID)},
		fieldVersion:   {N: aws.String("1")},
		fieldCreatedAt: {S: aws.String(now)},
		fieldUpdatedAt: {S: aws.String(now)},
		fieldEvents:    getOrNull(events),
		fieldPayload:   getOrNull(in.Payload),
	}

	input := dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(#id)"),
		ExpressionAttributeNames: map[string]*string{
			"#id": aws.String(fieldID),
		},
	}
	if _, err := d.api.PutItemWithContext(ctx, &input); err != nil {
		return errors.Wrapf(err, "unable to save record")
	}

	return nil
}

type updateIn struct {
	ID      string
	Version int
	Payload *dynamodb.AttributeValue
	Events  []Event
}

func (d dao) update(ctx context.Context, in updateIn) error {
	events, err := d.marshalEvents(in.Events)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal payload")
	}

	input := dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			fieldID: {S: aws.String(in.ID)},
		},
		ConditionExpression: aws.String("attribute_exists(#id) AND #version = :version"),
		UpdateExpression:    aws.String("SET #version = :version + :one, #uat = :uat, #payload = :payload, #events = :events"),
		ExpressionAttributeNames: map[string]*string{
			"#id":      aws.String(fieldID),
			"#uat":     aws.String(fieldUpdatedAt),
			"#payload": aws.String(fieldPayload),
			"#events":  aws.String(fieldEvents),
			"#version": aws.String(fieldVersion),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":one":     {N: aws.String("1")},
			":version": {N: aws.String(strconv.Itoa(in.Version))},
			":uat":     {S: aws.String(time.Now().Format(dateLayout))},
			":payload": getOrNull(in.Payload),
			":events":  getOrNull(events),
		},
	}
	if _, err := d.api.UpdateItemWithContext(ctx, &input); err != nil {
		return errors.Wrapf(err, "unable to save record")
	}

	return nil
}

type upsertIn struct {
	ID      string
	Version int
	Payload *dynamodb.AttributeValue
	Events  []Event
}

func (d dao) upsert(ctx context.Context, in upsertIn) error {
	if in.Version == 1 {
		return d.create(ctx, createIn{
			ID:      in.ID,
			Payload: in.Payload,
			Events:  in.Events,
		})
	} else {
		return d.update(ctx, updateIn{
			ID:      in.ID,
			Version: in.Version,
			Payload: in.Payload,
			Events:  in.Events,
		})
	}
}

func getOrNull(item *dynamodb.AttributeValue) *dynamodb.AttributeValue {
	if item == nil {
		return &dynamodb.AttributeValue{NULL: aws.Bool(true)}
	}

	return item
}

func unmarshalTime(item *dynamodb.AttributeValue) (time.Time, error) {
	if item == nil {
		return time.Time{}, errors.Errorf("invalid dynamodb time")
	}
	if item.NULL != nil && *item.NULL {
		return time.Time{}, nil
	}
	if item.S == nil {
		return time.Time{}, errors.Errorf("invalid dynamodb time")
	}

	t, err := time.Parse(dateLayout, *item.S)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "unable to parse time, %v", *item.S)
	}

	return t, nil
}

func unmarshalInt(item *dynamodb.AttributeValue) (int, error) {
	if item == nil || item.N == nil {
		return 0, errors.Errorf("invalid dynamodb number")
	}

	v, err := strconv.Atoi(*item.N)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to parse int, %v", *item.S)
	}

	return v, nil
}

func unmarshalString(item *dynamodb.AttributeValue) (string, error) {
	if item == nil || item.S == nil {
		return "", errors.Errorf("invalid dynamodb string")
	}

	return *item.S, nil
}
