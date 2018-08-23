package dynamodb

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
	"github.com/savaki/snapsource"
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

type record struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Payload   *dynamodb.AttributeValue
}

func (s *record) Unmarshal(item map[string]*dynamodb.AttributeValue) error {
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

	value := record{
		ID:        id,
		Version:   version,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Payload:   item[fieldPayload],
	}

	*s = value
	return nil
}

var errNotFound = errors.New("record not found")

type causer interface {
	Cause() error
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if err == errNotFound {
		return true
	}
	if v, ok := err.(causer); ok {
		return IsNotFound(v.Cause())
	}

	return false
}

type Store struct {
	api        dynamodbiface.DynamoDBAPI
	tableName  string
	serializer snapsource.Serializer
}

func (s Store) Load(ctx context.Context, id string, v interface{}) (snapsource.Meta, error) {
	r, ok, err := s.get(ctx, id)
	if err != nil {
		return snapsource.Meta{}, err
	}
	if !ok {
		return snapsource.Meta{}, errNotFound
	}

	if err := dynamodbattribute.Unmarshal(r.Payload, v); err != nil {
		return snapsource.Meta{}, errors.Wrapf(err, "unable to unmarshal record")
	}

	return snapsource.Meta{
		ID:        r.ID,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}, nil
}

func (s Store) get(ctx context.Context, id string) (record, bool, error) {
	input := dynamodb.GetItemInput{
		TableName:      aws.String(s.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			fieldID: {S: aws.String(id)},
		},
	}
	out, err := s.api.GetItemWithContext(ctx, &input)
	if err != nil {
		return record{}, false, errors.Wrapf(err, "unable to retrieve item from %v, id: %v", s.tableName)
	}
	if len(out.Item) == 0 {
		return record{}, false, nil
	}

	var v record
	if err := v.Unmarshal(out.Item); err != nil {
		return record{}, false, err
	}

	return v, true, nil
}

func (s Store) marshalEvents(events []snapsource.Event) (*dynamodb.AttributeValue, error) {
	var datum [][]byte
	for _, event := range events {
		data, err := s.serializer.MarshalEvent(event)
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
	Events  []snapsource.Event
}

func (s Store) create(ctx context.Context, in createIn) error {
	events, err := s.marshalEvents(in.Events)
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
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(#id)"),
		ExpressionAttributeNames: map[string]*string{
			"#id": aws.String(fieldID),
		},
	}
	if _, err := s.api.PutItemWithContext(ctx, &input); err != nil {
		return errors.Wrapf(err, "unable to save record")
	}

	return nil
}

type updateIn struct {
	ID      string
	Version int
	Payload *dynamodb.AttributeValue
	Events  []snapsource.Event
}

func (s Store) update(ctx context.Context, in updateIn) error {
	events, err := s.marshalEvents(in.Events)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal payload")
	}

	input := dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
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

	if _, err := s.api.UpdateItemWithContext(ctx, &input); err != nil {
		return errors.Wrapf(err, "unable to save record")
	}

	return nil
}

type upsertIn struct {
	ID      string
	Version int
	Payload *dynamodb.AttributeValue
	Events  []snapsource.Event
}

func (s Store) upsert(ctx context.Context, in upsertIn) error {
	if in.Version == 0 {
		return s.create(ctx, createIn{
			ID:      in.ID,
			Payload: in.Payload,
			Events:  in.Events,
		})
	} else {
		return s.update(ctx, updateIn{
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
