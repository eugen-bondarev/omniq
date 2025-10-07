package omniq

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type pgStorage[T any] struct {
	db      *sql.DB
	factory JobFactory[T]
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS jobs (
  id UUID NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  state JSONB NOT NULL DEFAULT '{}',
  type VARCHAR NOT NULL
)`

func NewPGStorage[T any](db *sql.DB, factory JobFactory[T]) *pgStorage[T] {
	_, err := db.Exec(createTableSQL)
	if err != nil {
		panic(err)
	}
	return &pgStorage[T]{db: db, factory: factory}
}

func (s *pgStorage[T]) Push(j Job[T], t time.Time) {
	id := uuid.New().String()
	state, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	_, err = s.db.Exec("INSERT INTO jobs (id, time, state, type) VALUES ($1, $2, $3, $4)", id, t, state, j.Type())
	if err != nil {
		panic(err)
	}
}

func (s *pgStorage[T]) Delete(j Job[T]) {
	_, err := s.db.Exec("DELETE FROM jobs WHERE id = $1", j.GetIDContainer().GetID())
	if err != nil {
		panic(err)
	}
}

func (s *pgStorage[T]) GetDue() []Job[T] {
	rows, err := s.db.Query("SELECT id, time, state, type FROM jobs WHERE time <= $1", time.Now())
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
