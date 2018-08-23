package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
	"github.com/savaki/snapsource"
)

type Handler struct {
	factory snapsource.Factory
	dao     Store
}

func (h *Handler) Apply(ctx context.Context, command snapsource.Command) ([]snapsource.Event, error) {
	state, ok, err := h.dao.get(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	v := h.factory(command.AggregateID())
	if ok {
		if err := dynamodbattribute.Unmarshal(state.Payload, v); err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal payload")
		}
	}

	handler := v.(snapsource.CommandHandler)
	events, err := handler.Apply(ctx, command)
	if err != nil {
		return nil, err
	}

	aggregate := v.(snapsource.Aggregate)
	for _, event := range events {
		if err := aggregate.On(event); err != nil {
			return nil, err
		}
	}

	item, err := dynamodbattribute.Marshal(aggregate)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal aggregate, %#v", aggregate)
	}

	in := upsertIn{
		ID:      command.AggregateID(),
		Version: state.Version,
		Payload: item,
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
	Serializer snapsource.Serializer
	Factory    snapsource.Factory
}

func New(config Config) (*Handler, error) {
	return &Handler{
		factory: config.Factory,
		dao: Store{
			api:        config.API,
			tableName:  config.TableName,
			serializer: config.Serializer,
		},
	}, nil
}
