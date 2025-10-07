package omniq

import (
	"log"
	"time"
)

type impl[T any] struct {
	storage SchedulerStorage[T]
}

func New[T any](storage SchedulerStorage[T]) *impl[T] {
	return &impl[T]{
		storage: storage,
	}
}

func NewWithDependencies[T any](storage SchedulerStorage[T]) *impl[T] {
	return &impl[T]{
		storage: storage,
	}
}

func (s *impl[T]) Listen(container T) {
	log.Println("Scheduler is running")
	for {
		for _, j := range s.storage.getDue() {
			j.Run(container)
			s.storage.delete(j)
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *impl[T]) ScheduleIn(j Job[T], d time.Duration) {
	s.storage.push(j, time.Now().Add(d))
}
