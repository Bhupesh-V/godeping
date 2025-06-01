package main

import (
	"flag"
	"fmt"
	"os"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
	ping "github.com/Bhupesh-V/godeping/ping"
	"github.com/Bhupesh-V/godeping/report"
	"github.com/Bhupesh-V/godeping/utils"
)

func main() {

	jsonOutput := flag.Bool("json", false, "Output results in JSON format (useful for scripting)")
	quiet := flag.Bool("quiet", false, "Suppress non-essential output (e.g., progress indicators)")
	sinceFlag := flag.String("since", "2y", "Consider dependencies as unmaintained if not updated since this duration (e.g. 1y, 6m, 2y3m)")
	flag.Usage = utils.GetUsageText()
	flag.Parse()

	// When JSON output is enabled, quiet mode is automatically turned on
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

	// Parse the duration from the since flag
	duration, err := utils.GetTimeDurationFromRelativeDate(*sinceFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid duration format for -since flag: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "Unable to parse go.mod file: %v\n", err)
		os.Exit(1)
	}

	if !*jsonOutput {
		fmt.Printf("Found %d dependencies in go.mod\n", len(moduleInfo.Requires))
		fmt.Printf("Module: %s\n", moduleInfo.ModuleName)
		fmt.Printf("Go Version: %s\n", moduleInfo.GoVersion)
	}

	// Always check for archived GitHub dependencies
	client := ping.NewClient()
	client.SetUnmaintainedDuration(duration)
	client.SetProgressCallback(utils.ProgressCallback(quiet))
	archivedResults := client.PingPackage(
		moduleInfo.Requires,
	)

	// Output the results using the appropriate format
	if *jsonOutput {
		report.OutputJSON(moduleInfo, archivedResults)
	} else {
		report.OutputText(moduleInfo, archivedResults)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <path-to-go-project>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}
