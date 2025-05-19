package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
	heartbeat "github.com/Bhupesh-V/godeping/ping"
)

func main() {
	// Define flags
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	quiet := flag.Bool("quiet", false, "Suppress progress output")

	// Parse flags
	flag.Parse()

	// When JSON output is enabled, quiet mode is automatically turned on
	// unless explicitly overridden
	if *jsonOutput {
		*quiet = true
	}

	// Check for the required positional argument
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: Path to Go project is required\n\n")
		printUsage()
		os.Exit(1)
	}

	projectPath := args[0]

	// Example usage of the flags and args
	if !*quiet {
		fmt.Printf("Analyzing Go project at: %s\n", projectPath)
	}

	// Parse the go.mod file
	moduleInfo, err := parser.ParseGoMod(projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !*jsonOutput {
		fmt.Printf("Found %d dependencies in go.mod\n", len(moduleInfo.Requires))
		fmt.Printf("Module: %s\n", moduleInfo.ModuleName)
		fmt.Printf("Go Version: %s\n", moduleInfo.GoVersion)
	}

	// Define a progress callback function
	progressCallback := func(dep string, status string) {
		if !*quiet {
			fmt.Printf("Analyzing: %-50s [%s]\n", dep, status)
		}
	}

	// Always check for archived GitHub dependencies
	client := heartbeat.NewClient()
	archivedResults := client.CheckArchivedDependenciesWithProgress(
		moduleInfo.Requires,
		progressCallback,
	)

	// Output the results
	if *jsonOutput {
		outputJSON(moduleInfo, archivedResults)
	} else {
		outputText(moduleInfo, archivedResults)
	}
}

func outputJSON(info *parser.ModuleInfo, repoStatus []heartbeat.RepoStatus) {
	// Count direct dependencies
	directDeps := 0

	// Get all direct dependencies
	var directDependencies []parser.Dependency
	for _, dep := range info.Requires {
		if !dep.Indirect {
			directDeps++
			directDependencies = append(directDependencies, dep)
		}
	}

	var archived []heartbeat.RepoStatus
	// Count archived dependencies
	for _, repo := range repoStatus {
		if repo.IsArchived {
			archived = append(archived, repo)
		}
	}

	type Output struct {
		Module               string                 `json:"module"`
		GoVersion            string                 `json:"goVersion"`
		TotalDependencies    int                    `json:"totalDependencies"`
		DirectDependencies   int                    `json:"directDependencies"`
		ArchivedDependencies []heartbeat.RepoStatus `json:"deadDirectDependencies"`
	}

	output := Output{
		Module:               info.ModuleName,
		GoVersion:            info.GoVersion,
		TotalDependencies:    len(info.Requires),
		DirectDependencies:   directDeps,
		ArchivedDependencies: archived,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func outputText(info *parser.ModuleInfo, archived []heartbeat.RepoStatus) {
	// Count direct dependencies
	directDeps := 0
	for _, dep := range info.Requires {
		if !dep.Indirect {
			directDeps++
		}
	}
	fmt.Printf("Direct Dependencies: %d\n", directDeps)

	// Print summary of archived repositories
	archivedCount := 0
	for _, repo := range archived {
		if repo.IsArchived {
			archivedCount++
		}
	}

	// Print archived GitHub dependencies if any
	if archivedCount > 0 {
		fmt.Println("\nArchived (Dead) Direct Dependencies:")
		for _, repo := range archived {
			if repo.IsArchived {
				fmt.Printf("%s\n", repo.ModulePath)
				if !repo.LastPublished.IsZero() {
					fmt.Printf(strings.Repeat(" ", 10))
					fmt.Printf("Last Published: %s\n", repo.LastPublished.Format("Jan 2, 2006"))
				}
			}
		}
	}

	// Print summary
	fmt.Println("\nSummary:")
	fmt.Printf("- Total Dependencies: %d\n", len(info.Requires))
	fmt.Printf("- Direct Dependencies: %d\n", directDeps)
	fmt.Printf("- Unmaintained Dependencies: %d\n", archivedCount)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <path-to-go-project>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}
