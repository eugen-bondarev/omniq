package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "generate":
		if err := runGenerate(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "init":
		if err := runInit(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if err := runAdd(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("omniq - Job queue code generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  omniq generate <jobs_directory>  Generate jobs_gen.go from job definitions")
	fmt.Println("  omniq init                       Initialize a jobs package in current directory")
	fmt.Println("  omniq add <job_name>             Add a new job to the jobs package")
	fmt.Println("  omniq help                       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  omniq generate ./jobs")
	fmt.Println("  omniq init")
	fmt.Println("  omniq add SendEmailJob")
}
