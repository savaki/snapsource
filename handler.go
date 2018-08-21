package esdb

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
)

type Handler struct {
	prototype reflect.Type
	dao       dao
}

func (h *Handler) Apply(ctx context.Context, command Command) ([]Event, error) {
	state, ok, err := h.dao.Get(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	v := reflect.New(h.prototype).Elem().Interface()
	if ok {
		if err := dynamodbattribute.Unmarshal(state.Payload, v); err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal payload")
		}
	}

	handler := v.(CommandHandler)
	events, err := handler.Apply(ctx, command)
	if err != nil {
		return nil, err
	}

	aggregate := v.(Aggregate)
	for _, event := range events {
		if err := aggregate.On(event); err != nil {
			return nil, err
		}
	}

	in := upsertIn{
		ID:      state.ID,
		Version: state.Version,
		Payload: nil,
		Events:  events,
	}
	if err := h.dao.upsert(ctx, in); err != nil {
		return nil, err
	}

	return events, nil
}

type Config struct {
	API        dynamodbiface.DynamoDBAPI
	TableName  string
	Serializer Serializer
	Prototype  reflect.Type
}

func New(config Config) (*Handler, error) {
	v := reflect.New(config.Prototype).Elem().Interface()
	if _, ok := v.(CommandHandler); !ok {
		return nil, errors.Errorf("%#v must implement CommandHandler", v)
	}
	if _, ok := v.(Aggregate); !ok {
		return nil, errors.Errorf("%#v must implement Aggregate", v)
	}

	return &Handler{
		prototype: config.Prototype,
		dao: dao{
			api:        config.API,
			tableName:  config.TableName,
			serializer: config.Serializer,
		},
	}, nil
}
