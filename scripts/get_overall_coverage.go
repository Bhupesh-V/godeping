//go:build ignore
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	// Run "go tool cover -func=coverage.out"
	cmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running go tool cover:", err)
		os.Exit(1)
	}

	// Scan through output to find the "total:" line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	totalRegex := regexp.MustCompile(`total:\s+\(statements\)\s+([\d.]+)%`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := totalRegex.FindStringSubmatch(line); matches != nil {
			fmt.Printf("%s%%\n",matches[1]) // Print only the coverage percentage
			return
		}
	}

	fmt.Println("Error: Could not find total coverage.")
	os.Exit(1)
}
