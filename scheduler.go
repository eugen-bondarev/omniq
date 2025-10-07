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
		jobs, err := s.storage.GetDue()
		if err != nil {
			log.Println("Error getting due jobs:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		for _, j := range jobs {
			j.Run(container)
			err = s.storage.Delete(j)
			if err != nil {
				log.Println("Error deleting job:", err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Scheduler[T]) ScheduleIn(j Job[T], d time.Duration) error {
	return s.storage.Push(j, time.Now().Add(d))
}
