package utils

import (
	"fmt"
	"os"
	"strings"
)

// ProgressCallback returns a function that can be used to report progress
func ProgressCallback(quiet *bool) func(string, string) {
	return func(dep string, status string) {
		if !*quiet {
			fmt.Printf("%-50s\n", dep)
			if status != "" {
				fmt.Print(strings.Repeat(" ", 50))
				fmt.Printf("[%s]\n", status)
			}
		}
	}
}

func GetUsageText() func() {
	return func() {
		fmt.Fprintf(os.Stdout, "godeping - Ping your Go project dependencies for aliveness (being maintained or not)\n")
		fmt.Fprintf(os.Stdout, "\nUsage:\n  %s [options] <path-to-go-project>\n\n", os.Args[0])
		fmt.Fprintln(os.Stdout, `
Examples:
========
Assuming you are in the root directory of your Go project:

	Run normally (with live progress):
		godeping .

	Run quietly (suppressing progress):
		godeping -quiet .

	Run quietly with JSON output:
		godeping -json .

	Check dependencies not updated in 6 months:
		godeping -since 6m .

	Check dependencies not updated in 1 year and 3 months:
		godeping -since 1y3m .

Support:
=======
	https://github.com/Bhupesh-V/godeping/issues`)
	}
}
