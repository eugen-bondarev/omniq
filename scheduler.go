package omniq

import (
	"log"
	"time"
)

type schedulerOptions struct {
	sleepDuration time.Duration
}

func newDefaultSchedulerOptions() schedulerOptions {
	return schedulerOptions{
		sleepDuration: 1 * time.Second,
	}
}

type schedulerOption func(*schedulerOptions)

func WithSleepDuration(d time.Duration) schedulerOption {
	return func(opts *schedulerOptions) {
		opts.sleepDuration = d
	}
}

type Scheduler[T any] struct {
	storage SchedulerStorage[T]
	options schedulerOptions
}

func New[T any](storage SchedulerStorage[T], opts ...schedulerOption) *Scheduler[T] {
	options := newDefaultSchedulerOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Scheduler[T]{
		storage: storage,
		options: options,
	}
}

func NewWithDependencies[T any](storage SchedulerStorage[T], opts ...schedulerOption) *Scheduler[T] {
	options := newDefaultSchedulerOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Scheduler[T]{
		storage: storage,
		options: options,
	}
}

func (s *Scheduler[T]) Listen(container T) {
	log.Println("Scheduler is running")
	for {
		jobs, err := s.storage.GetDue()
		if err != nil {
			log.Println("Error getting due jobs:", err)
			continue
		}

		for _, j := range jobs {
			j.Run(container)
			err = s.storage.Delete(j.GetIDContainer().GetID())
			if err != nil {
				log.Println("Error deleting job:", err)
			}
		}

		time.Sleep(s.options.sleepDuration)
	}
}

func (s *Scheduler[T]) ScheduleIn(j Job[T], d time.Duration) error {
	return s.storage.Push(j, time.Now().Add(d))
}
