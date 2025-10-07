package jobs

import (
	"github.com/eugen-bondarev/omniq"
	"github.com/eugen-bondarev/omniq/examples/postgres/deps"
)

type JobFactory struct{}

func (f *JobFactory) Instantiate(t string, id string, data map[string]any) omniq.Job[deps.Dependencies] {
	var j omniq.Job[deps.Dependencies]
	switch t {
	case "Job1":
		j = NewJob1(id, data)
	case "Job2":
		j = NewJob2(id, data)
	case "EmailJob":
		j = NewEmailJob(id, data)
	default:
		panic("Unknown job type: " + t)
	}
	return j
}
