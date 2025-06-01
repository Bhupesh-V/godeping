package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestProgressCallback(t *testing.T) {
	// Setup to capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Helper function to read captured output
	readOutput := func() string {
		outC := make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outC <- buf.String()
		}()

		w.Close()
		os.Stdout = oldStdout
		return <-outC
	}

	t.Run("normal output when quiet is false", func(t *testing.T) {
		// Reset pipe for this test
		r, w, _ = os.Pipe()
		os.Stdout = w

		quiet := false
		callback := ProgressCallback(&quiet)

		// Call the function
		callback("github.com/user/package", "updated")

		// Get captured output
		output := readOutput()

		// Check if output contains expected strings
		expectedDep := "github.com/user/package"
		expectedStatus := "[updated]"

		if !strings.Contains(output, expectedDep) {
			t.Errorf("Expected output to contain '%s', got: '%s'", expectedDep, output)
		}
		if !strings.Contains(output, expectedStatus) {
			t.Errorf("Expected output to contain '%s', got: '%s'", expectedStatus, output)
		}
	})

	t.Run("no output when quiet is true", func(t *testing.T) {
		// Reset pipe for this test
		r, w, _ = os.Pipe()
		os.Stdout = w

		quiet := true
		callback := ProgressCallback(&quiet)

		// Call the function
		callback("github.com/user/package", "updated")

		// Get captured output
		output := readOutput()

		if output != "" {
			t.Errorf("Expected no output when quiet is true, got: '%s'", output)
		}
	})

	t.Run("status message is optional", func(t *testing.T) {
		// Reset pipe for this test
		r, w, _ = os.Pipe()
		os.Stdout = w

		quiet := false
		callback := ProgressCallback(&quiet)

		// Call the function with empty status
		callback("github.com/user/package", "")

		// Get captured output
		output := readOutput()

		// Should contain the dependency name but no status brackets
		expectedDep := "github.com/user/package"
		unexpectedStr := "[]"

		if !strings.Contains(output, expectedDep) {
			t.Errorf("Expected output to contain '%s', got: '%s'", expectedDep, output)
		}
		if strings.Contains(output, unexpectedStr) {
			t.Errorf("Expected output not to contain empty brackets, got: '%s'", output)
		}
	})
}

func TestGetUsageText(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Get the usage function and execute it
	usageFunc := GetUsageText()
	usageFunc()

	// Close the write end of the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Test if the output contains essential parts
	expectedStrings := []string{
		"godeping - Ping your Go project dependencies",
		"Usage:",
		"Examples:",
		"godeping .",
		"godeping -quiet .",
		"godeping -json .",
		"godeping -since 6m .",
		"godeping -since 1y3m .",
		"Support:",
		"https://github.com/Bhupesh-V/godeping/issues",
	}

	for _, str := range expectedStrings {
		if !strings.Contains(output, str) {
			t.Errorf("Expected usage text to contain '%s', but it doesn't", str)
		}
	}
}
