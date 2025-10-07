package jobs

import (
	"github.com/eugen-bondarev/omniq"
	"github.com/eugen-bondarev/omniq/examples/postgres/deps"
)

// Jobs
func (j *Job1) Type() string {
	return "Job1"
}

func (j *Job2) Type() string {
	return "Job2"
}

func (j *EmailJob) Type() string {
	return "EmailJob"
}

func (j *Job1) GetIDContainer() *omniq.WithID {
	return &j.WithID
}

func (j *Job2) GetIDContainer() *omniq.WithID {
	return &j.WithID
}

func (j *EmailJob) GetIDContainer() *omniq.WithID {
	return &j.WithID
}

func NewJob1(id string, data map[string]any) *Job1 {
	return &Job1{
		WithID: omniq.WithID{
			ID: id,
		},
		MyData: data["MyData"].(string),
	}
}

func NewJob2(id string, data map[string]any) *Job2 {
	return &Job2{
		WithID: omniq.WithID{
			ID: id,
		},
		Answer: int(data["Answer"].(float64)),
	}
}

func NewEmailJob(id string, data map[string]any) *EmailJob {
	return &EmailJob{
		WithID: omniq.WithID{
			ID: id,
		},
		To:      data["To"].(string),
		Subject: data["Subject"].(string),
		Body:    data["Body"].(string),
	}
}

// Registry
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
