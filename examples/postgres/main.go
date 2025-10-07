package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/eugen-bondarev/omniq"
	"github.com/eugen-bondarev/omniq/examples/postgres/deps"
	"github.com/eugen-bondarev/omniq/examples/postgres/jobs"
	"github.com/eugen-bondarev/omniq/examples/postgres/services"

	_ "github.com/lib/pq"
)

type postgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func (pgConfig *postgresConfig) getConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Password, pgConfig.DBName)
}

func main() {
	pgConfig := postgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "postgres",
	}

	db, err := sql.Open("postgres", pgConfig.getConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create jobs
	j1 := &jobs.Job1{MyData: "Hello"}
	j2 := &jobs.Job2{Answer: 42}
	emailJob := &jobs.EmailJob{
		To:      "user@example.com",
		Subject: "Welcome!",
		Body:    "<h1>Hello from the job scheduler!</h1>",
	}

	factory := &jobs.JobFactory{}

	pgStorage := omniq.NewPGStorage(db, factory)
	// jsonStorage := scheduler.NewJSONStorage[deps.Dependencies]("scheduler_state.json")

	container := deps.Dependencies{
		Mailer: &services.MockSMTPService{},
	}

	// Create scheduler with dependencies
	s := omniq.NewWithDependencies(pgStorage)
	// s := scheduler.NewWithDependencies(scheduler.NewJSONStorage[deps.Dependencies]("scheduler_state.json"))

	// Schedule jobs
	s.ScheduleIn(j1, 2*time.Second)
	s.ScheduleIn(j2, 1*time.Second)
	s.ScheduleIn(emailJob, 3*time.Second)

	go s.Listen(container)

	for {
		time.Sleep(1 * time.Second)
	}
}
