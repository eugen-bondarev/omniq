# Omniq - A simple, fast, database-agnostic async job scheduler for cloud-native Go apps

## Motivation

Yes, Redis exists, and there are tons of solutions for this problem based on Redis. However, I oftentimes find myself in simple projects where I only use a single Postgres instance, or even SQLite in a bucket mounted into the app. Furthermore, I always work in dockerized environments where the app is stateless, which means you can't use `timer.AfterFunc` to schedule jobs - they'll simply get lost after the app is restarted because of a deployment.

### How Omniq addresses these problems

- Omniq is database-agnostic - just implement the `SchedulerStorage` interface and you're good to go

```go
type SchedulerStorage[TDeps any] interface {
	Push(j Job[TDeps], t time.Time) error
	Delete(j Job[TDeps]) error
	GetDue() ([]Job[TDeps], error)
}
```

To get an inspiration, check out `json_storage.go` and `pg_storage.go` for example implementations. Or you can use existing implementations.

- Codegen makes Omniq fast because you don't have to rely on reflection

- Because of Codegen you get a very simple and clear interface to work with:

```go
type TestJob struct {
	omniq.WithID
	MyData string
}

func (j *TestJob) Run(d deps.Dependencies) {
	log.Println("TestJob is running", j.MyData)
}
```

## How to use in your project

Initialize omniq in your project:

```
go run -mod=mod github.com/eugen-bondarev/omniq/cmd/omniq init
```

Add a new job:

```
go run -mod=mod github.com/eugen-bondarev/omniq/cmd/omniq add {job name}
```

Generate the `jobs_gen.go` file:

```
go generate ./jobs
```

Somewhere in your app's initialization code (example with postgres backend):

```go
db, err := sql.Open("postgres", pgConfig.getConnectionString())
if err != nil {
    log.Fatal(err)
}

factory := &jobs.JobFactory{}

pg, err := omniq.NewPGStorage(db, factory, omniq.WithTableName("lorem_ipsum"))
if err != nil {
    log.Fatal(err)
}
scheduler := omniq.NewWithDependencies(pg, omniq.WithSleepDuration(1*time.Second))

go scheduler.Listen(jobs.Dependencies{})
```

Some time later when some event occurs:

```go
sayHiJob := &jobs.SayHiJob{
    Name: "John Doe",
}

scheduler.ScheduleIn(sayHiJob, 60*time.Minute)
```

Check out the [examples](https://github.com/eugen-bondarev/omniq/tree/main/examples) for more details.
