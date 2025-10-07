package omniq

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type pgStorageOptions struct {
	tableName string
}

type pgStorageOption func(*pgStorageOptions)

func WithTableName(tableName string) pgStorageOption {
	return func(opts *pgStorageOptions) {
		opts.tableName = tableName
	}
}

type pgStorage[T any] struct {
	db      *sql.DB
	factory JobFactory[T]
	options pgStorageOptions
}

func newDefaultPGStorageOptions() pgStorageOptions {
	return pgStorageOptions{tableName: "omniq_jobs"}
}

func NewPGStorage[T any](db *sql.DB, factory JobFactory[T], opts ...pgStorageOption) *pgStorage[T] {
	options := newDefaultPGStorageOptions()
	for _, opt := range opts {
		opt(&options)
	}

	s := &pgStorage[T]{db: db, factory: factory, options: options}
	s.createTable()
	return s
}

func (s *pgStorage[T]) createTable() {
	cmd := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id UUID NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  state JSONB NOT NULL DEFAULT '{}',
  type VARCHAR NOT NULL
)`, s.options.tableName)

	_, err := s.db.Exec(cmd)
	if err != nil {
		panic(err)
	}
}

func (s *pgStorage[T]) Push(j Job[T], t time.Time) {
	id := uuid.New().String()
	state, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	_, err = s.db.Exec("INSERT INTO "+s.options.tableName+" (id, time, state, type) VALUES ($1, $2, $3, $4)", id, t, state, j.Type())
	if err != nil {
		panic(err)
	}
}

func (s *pgStorage[T]) Delete(j Job[T]) {
	_, err := s.db.Exec("DELETE FROM "+s.options.tableName+" WHERE id = $1", j.GetIDContainer().GetID())
	if err != nil {
		panic(err)
	}
}

func (s *pgStorage[T]) GetDue() []Job[T] {
	rows, err := s.db.Query("SELECT id, time, state, type FROM "+s.options.tableName+" WHERE time <= $1", time.Now())
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	due := []Job[T]{}
	for rows.Next() {
		var id string
		var t time.Time
		var stateEncoded json.RawMessage
		var typ string
		err = rows.Scan(&id, &t, &stateEncoded, &typ)
		if err != nil {
			panic(err)
		}
		state := map[string]any{}
		err = json.Unmarshal(stateEncoded, &state)
		if err != nil {
			panic(err)
		}
		j := s.factory.Instantiate(typ, id, state)
		due = append(due, j)
	}

	return due
}
