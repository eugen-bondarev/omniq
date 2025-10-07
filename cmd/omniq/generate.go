package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// runGenerate handles the generate command
func runGenerate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("generate command requires a jobs directory argument")
	}

	jobsDir := args[0]

	// Parse all job files in the directory
	jobs, packageName, depType, depImport, err := parseJobsDirectory(jobsDir)
	if err != nil {
		return fmt.Errorf("parsing jobs directory: %v", err)
	}

	if len(jobs) == 0 {
		return fmt.Errorf("no job structs found in %s", jobsDir)
	}

	// Generate the code
	data := GenerationData{
		Package:   packageName,
		Jobs:      jobs,
		DepType:   depType,
		DepImport: depImport,
	}

	tmpl, err := template.New("generated").Parse(generateTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %v", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting generated code: %v", err)
	}

	// Write to jobs_gen.go in the same directory
	outputFile := filepath.Join(jobsDir, "jobs_gen.go")

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return fmt.Errorf("writing generated file: %v", err)
	}

	fmt.Printf("Generated %s\n", outputFile)
	return nil
}
