package omniq

import (
	"time"
)

type SchedulerStorage[TDeps any] interface {
	push(j Job[TDeps], t time.Time)
	delete(j Job[TDeps])
	getDue() []Job[TDeps]
}
