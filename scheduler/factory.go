package scheduler

import "github.com/eugen-bondarev/omniq/job"

type JobFactory[T any] interface {
	Instantiate(t string, id string, data map[string]any) job.Job[T]
}
