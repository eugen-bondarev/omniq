package omniq

import (
	"time"
)

type SchedulerStorage[TDeps any] interface {
	Push(j Job[TDeps], t time.Time)
	Delete(j Job[TDeps])
	GetDue() []Job[TDeps]
}
