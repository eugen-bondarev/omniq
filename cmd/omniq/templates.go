package main

const initJobTemplate = `package jobs

import (
	"log"

	"github.com/eugen-bondarev/omniq"
)

type SayHiJob struct {
	omniq.WithID
	Name string
}

func (j *SayHiJob) Run(d Dependencies) {
	log.Println("Hi,", j.Name)
}
`

const exampleDepsTemplate = `package jobs

type Dependencies struct {
}
`

const addJobTemplate = `package jobs

import (
	"log"

	"github.com/eugen-bondarev/omniq"
)

type {{.JobName}} struct {
	omniq.WithID
	// Add your job fields here
}

func (j *{{.JobName}}) Run({{.RunParams}}) {
	log.Println("{{.JobName}} is running")
	// Add your job logic here
}
`

const generateTemplate = `package {{.Package}}

import (
	"github.com/eugen-bondarev/omniq"
{{if .DepImport}}	"{{.DepImport}}"{{end}}
)

// Jobs
{{range .Jobs}}func (j *{{.Name}}) Type() string {
	return "{{.Name}}"
}

{{end}}{{range .Jobs}}func (j *{{.Name}}) GetIDContainer() *omniq.WithID {
	return &j.WithID
}

{{end}}{{range .Jobs}}func New{{.Name}}(id omniq.JobID, data map[string]any) *{{.Name}} {
	return &{{.Name}}{
		WithID: omniq.WithID{
			ID: id,
		},{{range .Fields}}
		{{.Name}}: data["{{.Name}}"].({{.Type}}),{{end}}
	}
}

{{end}}// Registry
type JobFactory struct{}

func (f *JobFactory) Instantiate(t string, id omniq.JobID, data map[string]any) omniq.Job[{{.DepType}}] {
	var j omniq.Job[{{.DepType}}]
	switch t {
{{range .Jobs}}	case "{{.Name}}":
		j = New{{.Name}}(id, data)
{{end}}	default:
		panic("Unknown job type: " + t)
	}
	return j
}
`

const generateFileDirective = `//go:generate sh -c "cd .. && go run github.com/eugen-bondarev/omniq/cmd/omniq generate jobs"

package jobs`
