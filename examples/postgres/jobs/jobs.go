package jobs

import (
	"log"

	"github.com/eugen-bondarev/omniq"
	"github.com/eugen-bondarev/omniq/examples/postgres/deps"
)

type Job1 struct {
	omniq.WithID
	MyData string
}

func (j *Job1) Run(container deps.Dependencies) {
	log.Println("Job1 is running", j.MyData)
}

type Job2 struct {
	omniq.WithID
	Answer int
}

func (j *Job2) Run(container deps.Dependencies) {
	log.Println("Job2 is running", j.Answer)
}

// EmailJob demonstrates dependency injection for SMTP service
type EmailJob struct {
	omniq.WithID
	To      string
	Subject string
	Body    string
}

func (j *EmailJob) Run(container deps.Dependencies) {
	err := container.Mailer.SendEmail(j.To, j.Subject, j.Body)
	if err != nil {
		log.Println("SMTP service not available")
		return
	}

	log.Printf("Email sent successfully to %s", j.To)
}
