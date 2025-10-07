package omniq

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

type jsonEntry struct {
	ID    string
	Time  time.Time
	State any
	Type  string
}

type jsonStorage[T any] struct {
	fileName string
	factory  JobFactory[T]
}

func NewJSONStorage[T any](fileName string, factory JobFactory[T]) *jsonStorage[T] {
	return &jsonStorage[T]{fileName: fileName, factory: factory}
}

func (s *jsonStorage[T]) Push(j Job[T], t time.Time) {
	content, err := os.ReadFile(s.fileName)
	if err != nil {
		panic(err)
	}

	entries := []jsonEntry{}
	json.Unmarshal(content, &entries)

	j.GetIDContainer().SetID(uuid.New().String())
	entries = append(entries, jsonEntry{ID: j.GetIDContainer().GetID(), Time: t, State: j, Type: j.Type()})
	content, err = json.Marshal(entries)
	if err != nil {
		panic(err)
	}
	os.WriteFile(s.fileName, content, 0644)
}

func (s *jsonStorage[T]) Delete(j Job[T]) {
	content, err := os.ReadFile(s.fileName)
	if err != nil {
		panic(err)
	}
	entries := []jsonEntry{}
	json.Unmarshal(content, &entries)
	for t, e := range entries {
		if e.ID == j.GetIDContainer().GetID() {
			entries = append(entries[:t], entries[t+1:]...)
			break
		}
	}
	content, err = json.Marshal(entries)
	if err != nil {
		panic(err)
	}
	os.WriteFile(s.fileName, content, 0644)
}

func (s *jsonStorage[T]) GetDue() []Job[T] {
	now := time.Now()
	due := []Job[T]{}
	entries := []jsonEntry{}
	content, err := os.ReadFile(s.fileName)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(content, &entries)
	for _, e := range entries {
		if e.Time.Before(now) {
			j := s.factory.Instantiate(e.Type, e.ID, e.State.(map[string]any))
			due = append(due, j)
		}
	}
	return due
}
