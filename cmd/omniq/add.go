package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// runAdd handles the add command
func runAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("add command requires a job name argument")
	}

	jobName := args[0]

	// Validate job name
	if !strings.HasSuffix(jobName, "Job") {
		jobName += "Job"
	}

	if !isValidGoIdentifier(jobName) {
		return fmt.Errorf("invalid job name: %s (must be a valid Go identifier)", jobName)
	}

	// Check if jobs directory exists
	jobsDir := "jobs"
	if _, err := os.Stat(jobsDir); os.IsNotExist(err) {
		return fmt.Errorf("jobs directory does not exist. Run 'omniq init' first")
	}

	// Detect existing dependency type
	depType, depImport, err := detectExistingJobDependency(jobsDir)
	if err != nil {
		return fmt.Errorf("detecting existing job dependency: %v", err)
	}

	// Prepare template data
	var runParams string
	if depType == "struct{}" {
		runParams = ""
	} else {
		if depImport != "" {
			// Extract package name from import path
			parts := strings.Split(depImport, "/")
			pkgName := parts[len(parts)-1]
			runParams = fmt.Sprintf("d %s.%s", pkgName, strings.TrimPrefix(depType, pkgName+"."))
		} else {
			runParams = fmt.Sprintf("d %s", depType)
		}
	}

	data := AddJobData{
		JobName:   jobName,
		RunParams: runParams,
	}

	// Parse and execute template
	tmpl, err := template.New("addJob").Parse(addJobTemplate)
	if err != nil {
		return fmt.Errorf("parsing add job template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing add job template: %v", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting generated code: %v", err)
	}

	// Create job file
	filename := strings.ToLower(strings.ReplaceAll(jobName, "Job", "_job")) + ".go"
	if !strings.HasSuffix(filename, "_job.go") {
		filename = strings.TrimSuffix(filename, ".go") + "_job.go"
	}

	jobFile := filepath.Join(jobsDir, filename)

	// Check if file already exists
	if _, err := os.Stat(jobFile); err == nil {
		return fmt.Errorf("job file %s already exists", jobFile)
	}

	if err := os.WriteFile(jobFile, formatted, 0644); err != nil {
		return fmt.Errorf("creating job file: %v", err)
	}

	fmt.Printf("Added job %s to %s\n", jobName, jobFile)
	fmt.Printf("\nTo regenerate the boilerplate code, run:\n")
	fmt.Printf("  go generate ./jobs\n")

	return nil
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be a letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}
