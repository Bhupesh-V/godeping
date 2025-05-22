package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
	ping "github.com/Bhupesh-V/godeping/ping"
)

// OutputJSON prints the results in JSON format
func OutputJSON(info *parser.ModuleInfo, repoStatus []ping.RepoStatus) {
	// Get all direct dependencies
	var directDependencies []parser.Dependency
	for _, dep := range info.Requires {
		if !dep.Indirect {
			directDependencies = append(directDependencies, dep)
		}
	}

	var archived []ping.RepoStatus
	// Count archived dependencies
	for _, repo := range repoStatus {
		if repo.IsArchived {
			archived = append(archived, repo)
		}
	}

	type Output struct {
		Module               string            `json:"module"`
		GoVersion            string            `json:"goVersion"`
		TotalDependencies    int               `json:"totalDependencies"`
		DirectDependencies   int               `json:"directDependencies"`
		ArchivedDependencies []ping.RepoStatus `json:"deadDirectDependencies"`
	}

	output := Output{
		Module:               info.ModuleName,
		GoVersion:            info.GoVersion,
		TotalDependencies:    len(info.Requires),
		DirectDependencies:   len(directDependencies),
		ArchivedDependencies: archived,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

// OutputText prints the results in human-readable text format
func OutputText(info *parser.ModuleInfo, archived []ping.RepoStatus) {
	// Count direct dependencies
	directDeps := 0
	for _, dep := range info.Requires {
		if !dep.Indirect {
			directDeps++
		}
	}

	// Print summary of archived repositories
	archivedCount := 0
	for _, repo := range archived {
		if repo.IsArchived {
			archivedCount++
		}
	}

	// Print archived dependencies if any
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

// ProgressCallback returns a function that can be used to report progress
func ProgressCallback(quiet *bool) func(string, string) {
	return func(dep string, status string) {
		if !*quiet {
			fmt.Printf("%-50s\n", dep)
			if status != "" {
				fmt.Printf(strings.Repeat(" ", 50))
				fmt.Printf("[%s]\n", status)
			}
		}
	}
}
