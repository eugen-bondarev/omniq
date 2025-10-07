package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// runInit handles the init command
func runInit(args []string) error {
	// Create jobs directory
	jobsDir := "jobs"
	if err := os.MkdirAll(jobsDir, 0755); err != nil {
		return fmt.Errorf("creating jobs directory: %v", err)
	}

	// Create the generate.go file
	generateFile := filepath.Join(jobsDir, "generate.go")
	if err := os.WriteFile(generateFile, []byte(generateFileDirective), 0644); err != nil {
		return fmt.Errorf("creating generate.go: %v", err)
	}

	// Create the example job file
	exampleFile := filepath.Join(jobsDir, "say_hi_job.go")
	if err := os.WriteFile(exampleFile, []byte(initJobTemplate), 0644); err != nil {
		return fmt.Errorf("creating example job file: %v", err)
	}

	exampleDepsFile := filepath.Join(jobsDir, "deps.go")
	if err := os.WriteFile(exampleDepsFile, []byte(exampleDepsTemplate), 0644); err != nil {
		return fmt.Errorf("creating example deps file: %v", err)
	}

	fmt.Printf("Initialized jobs package in %s/\n", jobsDir)
	fmt.Printf("Created files:\n")
	fmt.Printf("  - %s\n", generateFile)
	fmt.Printf("  - %s\n", exampleFile)
	fmt.Printf("\nTo generate the boilerplate code, run:\n")
	fmt.Printf("  go generate ./jobs\n")

	return nil
}
