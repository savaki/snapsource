package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/savaki/snapsource"
)

type snapshot struct {
	ID        string `gorm:"primary_key"`
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Payload   string
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

const (
	createSnapshot = `create table if not exists %v (
	id         varchar(255) primary key,
	version    int not null,
	created_at datetime not null,
	updated_at datetime not null,
	payload    json
)`

	createEvents = `create table if not exists %v (
	id      varchar(255) not null,
	version int not null,
	at      datetime not null,
	data    varbinary(64000),
	primary key (id, version)
)`

	createIndex = `create index %v_idx on %v (at)`

	loadSQL = `select id, version, created_at, updated_at, payload from %v where id = ?;`

	createSQL = `insert into %v (id, version, created_at, updated_at, payload) values (?, ?, ?, ?, ?);`

	updateSQL = `update %v set version = ?, updated_at = ?, payload = ? where id = ?;`

	eventsSQL = `insert into %v (id, version, at, data) values (?, ?, ?, ?);`
)

type dao struct {
	config    Config
	loadSQL   string
	createSQL string
	updateSQL string
	eventsSQL string
}

func newDAO(config Config) dao {
	return dao{
		config:    config,
		loadSQL:   fmt.Sprintf(loadSQL, config.SnapshotTable),
		createSQL: fmt.Sprintf(createSQL, config.SnapshotTable),
		updateSQL: fmt.Sprintf(updateSQL, config.SnapshotTable),
		eventsSQL: fmt.Sprintf(eventsSQL, config.EventsTable),
	}
}

func (d dao) Load(ctx context.Context, id string, v interface{}) (snapsource.Meta, error) {
	var s snapshot
	row := d.config.DB.QueryRow(d.loadSQL, id)

	if err := row.Scan(&s.ID, &s.Version, &s.CreatedAt, &s.UpdatedAt, &s.Payload); err != nil {
		return snapsource.Meta{}, errors.Wrapf(err, "unable to read")
	}

	if err := json.NewDecoder(strings.NewReader(s.Payload)).Decode(v); err != nil {
		return snapsource.Meta{}, errors.Wrapf(err, "unable to decode record")
	}

	return snapsource.Meta{
		ID:        id,
		Version:   s.Version,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

func (d dao) Save(ctx context.Context, meta snapsource.Meta, v interface{}, events ...snapsource.Event) (err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal payload as JSON")
	}

	var eventArgs [][]interface{}
	for _, event := range events {
		data, err := d.config.Serializer.MarshalEvent(event)
		if err != nil {
			return errors.Wrapf(err, "unable to save event")
		}

		// insert into %v (id, version, at, data) values (?, ?, ?, ?)
		eventArgs = append(eventArgs, []interface{}{
			meta.ID,
			event.EventVersion(),
			event.EventAt(),
			data,
		})
	}

	tx, err := d.config.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to begin sql transaction")
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// create the record
	//
	if meta.Version == 1 {
		args := []interface{}{
			// id, version, created_at, updated_at, payload
			meta.ID, meta.Version, meta.CreatedAt, meta.UpdatedAt, data,
		}

		if _, err := tx.ExecContext(ctx, d.createSQL, args...); err != nil {
			return errors.Wrapf(err, "unable to insert into %v", d.config.SnapshotTable)
		}
	} else {
		// update %v set version = ?, updated_at = ?, payload = ? where id = ?
		args := []interface{}{
			meta.Version, meta.UpdatedAt, data, meta.ID,
		}

		result, err := tx.ExecContext(ctx, d.updateSQL, args...)
		if err != nil {
			return errors.Wrapf(err, "unable to update %v", d.config.SnapshotTable)
		}
		if n, err := result.RowsAffected(); err != nil || n == 0 {
			return errors.Errorf("no records updated")
		}
		n, err := result.RowsAffected()
		fmt.Println("result.RowsAffected", n, err)
	}

	// insert events
	//
	for _, args := range eventArgs {
		if _, err := tx.ExecContext(ctx, d.eventsSQL, args...); err != nil {
			return errors.Wrapf(err, "unable to insert into %v", d.config.SnapshotTable)
		}
	}

	return nil
}

func (d dao) AutoMigrate(ctx context.Context) error {
	sql := fmt.Sprintf(createSnapshot, d.config.SnapshotTable)
	if _, err := d.config.DB.ExecContext(ctx, sql); err != nil {
		return errors.Wrapf(err, "unable to create table, %v", d.config.SnapshotTable)
	}

	sql = fmt.Sprintf(createEvents, d.config.EventsTable)
	if _, err := d.config.DB.ExecContext(ctx, sql); err != nil {
		return errors.Wrapf(err, "unable to create table, %v", d.config.EventsTable)
	}

	sql = fmt.Sprintf(createIndex, d.config.EventsTable, d.config.EventsTable)
	if _, err := d.config.DB.ExecContext(ctx, sql); err != nil {
		if v, ok := err.(*mysql.MySQLError); !ok || v.Number != 1061 {
			return errors.Wrapf(err, "unable to create index on table, %v", d.config.EventsTable)
		}
	}

	return nil
}
