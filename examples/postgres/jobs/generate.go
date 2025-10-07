//go:generate sh -c "cd ../../../ && go build -o omniq-gen ./cmd/omniq && ./omniq-gen examples/postgres/jobs/jobs.go && rm omniq-gen"
package jobs
