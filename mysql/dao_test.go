package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/savaki/snapsource"
	"github.com/savaki/snapsource/dynamo/testdata/pb"
	"github.com/tj/assert"
)

type dbConfig struct {
	Username    string
	Password    string
	Hostname    string
	Port        string
	Database    string
	TxIsolation string
}

func connectString(cfg dbConfig) string {
	isolation := cfg.TxIsolation
	if isolation == "" {
		isolation = "READ-COMMITTED"
	}

	return fmt.Sprintf(`%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local&tx_isolation="%v"`,
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
		isolation,
	)
}

type Person struct {
	Name string
}

func TestDAO(t *testing.T) {
	str := connectString(dbConfig{
		Username: "snapsource",
		Password: "password",
		Hostname: "127.0.0.1",
		Port:     "3306",
		Database: "snapsource",
	})
	db, err := sql.Open("mysql", str)
	assert.Nil(t, err)
	defer db.Close()

	ctx := context.Background()
	config := Config{
		DB:            db,
		SnapshotTable: "users",
		EventsTable:   "users_events",
		Serializer:    pb.NewSerializer(),
	}

	dao := newDAO(config)
	err = dao.AutoMigrate(ctx)
	assert.Nil(t, err)

	meta := snapsource.Meta{
		ID:        strconv.FormatInt(time.Now().UnixNano(), 36),
		Version:   1,
		UpdatedAt: time.Now().Add(time.Second).Round(time.Second),
		CreatedAt: time.Now().Round(time.Second),
	}

	b := pb.NewBuilder(meta.ID, meta.Version)
	b.EmailUpdated("old", "new")

	found := Person{Name: "Joe"}
	err = dao.Save(ctx, meta, found, b.Events...)
	assert.Nil(t, err)

	var got Person
	m, err := dao.Load(ctx, meta.ID, &got)
	assert.Nil(t, err)
	assert.Equal(t, meta, m)
	assert.Equal(t, found, got)
}
