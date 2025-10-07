package omniq

import (
	"time"
)

type SchedulerStorage[TDeps any] interface {
	Push(j Job[TDeps], t time.Time) error
	Delete(j Job[TDeps]) error
	GetDue() ([]Job[TDeps], error)
}
