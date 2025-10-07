package scheduler

import (
	"time"

	"github.com/eugen-bondarev/omniq/job"
)

type SchedulerStorage[TDeps any] interface {
	push(j job.Job[TDeps], t time.Time)
	delete(j job.Job[TDeps])
	getDue() []job.Job[TDeps]
}
