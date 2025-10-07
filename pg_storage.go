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

func newDefaultPGStorageOptions() pgStorageOptions {
	return pgStorageOptions{tableName: "omniq_jobs"}
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

func NewPGStorage[T any](db *sql.DB, factory JobFactory[T], opts ...pgStorageOption) (*pgStorage[T], error) {
	options := newDefaultPGStorageOptions()
	for _, opt := range opts {
		opt(&options)
	}

	s := &pgStorage[T]{db: db, factory: factory, options: options}
	err := s.createTable()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *pgStorage[T]) createTable() error {
	cmd := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id UUID NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  state JSONB NOT NULL DEFAULT '{}',
  type VARCHAR NOT NULL
)`, s.options.tableName)

	_, err := s.db.Exec(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (s *pgStorage[T]) Push(j Job[T], t time.Time) error {
	id := uuid.New().String()
	state, err := json.Marshal(j)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO "+s.options.tableName+" (id, time, state, type) VALUES ($1, $2, $3, $4)", id, t, state, j.Type())
	if err != nil {
		return err
	}
	return nil
}

func (s *pgStorage[T]) Delete(id JobID) error {
	_, err := s.db.Exec("DELETE FROM "+s.options.tableName+" WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *pgStorage[T]) GetDue() ([]Job[T], error) {
	rows, err := s.db.Query("SELECT id, time, state, type FROM "+s.options.tableName+" WHERE time <= $1", time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	due := []Job[T]{}
	for rows.Next() {
		var id JobID
		var t time.Time
		var stateEncoded json.RawMessage
		var typ string
		err = rows.Scan(&id, &t, &stateEncoded, &typ)
		if err != nil {
			return nil, err
		}
		state := map[string]any{}
		err = json.Unmarshal(stateEncoded, &state)
		if err != nil {
			return nil, err
		}
		j := s.factory.Instantiate(typ, id, state)
		due = append(due, j)
	}

	return due, nil
}
