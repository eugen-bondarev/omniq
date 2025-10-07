package omniq

import (
	"log"
	"time"
)

type Scheduler[T any] struct {
	storage SchedulerStorage[T]
}

func New[T any](storage SchedulerStorage[T]) *Scheduler[T] {
	return &Scheduler[T]{
		storage: storage,
	}
}

func NewWithDependencies[T any](storage SchedulerStorage[T]) *Scheduler[T] {
	return &Scheduler[T]{
		storage: storage,
	}
}

func (s *Scheduler[T]) Listen(container T) {
	log.Println("Scheduler is running")
	for {
		for _, j := range s.storage.GetDue() {
			j.Run(container)
			s.storage.Delete(j)
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Scheduler[T]) ScheduleIn(j Job[T], d time.Duration) {
	s.storage.Push(j, time.Now().Add(d))
}
