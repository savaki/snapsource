package mysql

import (
	"context"
	"database/sql"

	"github.com/savaki/snapsource"
)

type Config struct {
	DB            *sql.DB
	SnapshotTable string
	EventsTable   string
	Serializer    snapsource.Serializer
}

type Handler struct {
	config Config
}

func (h *Handler) Apply(ctx context.Context, command snapsource.Command) ([]snapsource.Event, error) {
	return nil, nil
}

func New(config Config) (*Handler, error) {
	return &Handler{
		config: config,
	}, nil
}
