package jobs

import "github.com/eugen-bondarev/omniq/job"

func (j *Job1) Type() string {
	return "Job1"
}

func (j *Job2) Type() string {
	return "Job2"
}

func (j *EmailJob) Type() string {
	return "EmailJob"
}

func (j *Job1) GetIDContainer() *job.WithID {
	return &j.WithID
}

func (j *Job2) GetIDContainer() *job.WithID {
	return &j.WithID
}

func (j *EmailJob) GetIDContainer() *job.WithID {
	return &j.WithID
}

func NewJob1(id string, data map[string]any) *Job1 {
	return &Job1{
		WithID: job.WithID{
			ID: id,
		},
		MyData: data["MyData"].(string),
	}
}

func NewJob2(id string, data map[string]any) *Job2 {
	return &Job2{
		WithID: job.WithID{
			ID: id,
		},
		Answer: int(data["Answer"].(float64)),
	}
}

func NewEmailJob(id string, data map[string]any) *EmailJob {
	return &EmailJob{
		WithID: job.WithID{
			ID: id,
		},
		To:      data["To"].(string),
		Subject: data["Subject"].(string),
		Body:    data["Body"].(string),
	}
}
