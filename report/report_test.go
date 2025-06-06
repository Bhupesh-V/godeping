package report

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	parser "github.com/Bhupesh-V/godeping/parsers/modfile"
	"github.com/Bhupesh-V/godeping/ping"
)

// setupTestModuleInfo creates a test ModuleInfo structure
func setupTestModuleInfo() parser.ModuleInfo {
	return parser.ModuleInfo{
		ModuleName: "github.com/example/testmodule",
		GoVersion:  "1.18",
		Requires: []parser.Dependency{
			{Path: "github.com/active/repo", Version: "v1.0.0"},
			{Path: "github.com/archived/repo", Version: "v2.0.0"},
		},
	}
}

// setupRepoStatusResults creates test repo status results
func setupRepoStatusResults() []ping.RepoStatus {
	return []ping.RepoStatus{
		{ModulePath: "github.com/active/repo", IsArchived: false},
		{ModulePath: "github.com/archived/repo", IsArchived: true},
	}
}

func TestOutputJSON(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	moduleInfo := setupTestModuleInfo()
	repoResults := setupRepoStatusResults()

	// Call the function
	OutputJSON(&moduleInfo, repoResults)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Print output for debugging
	t.Logf("JSON Output: %s", output)

	// Verify the JSON output
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the structure and content based on the actual JSON structure
	if result["module"] != moduleInfo.ModuleName {
		t.Errorf("Expected module name %s, got %s", moduleInfo.ModuleName, result["module"])
	}

	if result["goVersion"] != moduleInfo.GoVersion {
		t.Errorf("Expected Go version %s, got %s", moduleInfo.GoVersion, result["goVersion"])
	}

	// Check totalDependencies
	if result["totalDependencies"] != float64(len(moduleInfo.Requires)) {
		t.Errorf("Expected totalDependencies %d, got %v", len(moduleInfo.Requires), result["totalDependencies"])
	}

	// Check directDependencies
	expectedDirectDeps := 0
	for _, dep := range moduleInfo.Requires {
		if !dep.Indirect {
			expectedDirectDeps++
		}
	}
	if result["directDependencies"] != float64(expectedDirectDeps) {
		t.Errorf("Expected directDependencies %d, got %v", expectedDirectDeps, result["directDependencies"])
	}

	// Check archived dependencies
	archivedDeps, ok := result["deadDirectDependencies"].([]interface{})
	if !ok {
		t.Fatal("deadDirectDependencies not found or not an array in JSON output")
	}

	// We expect only the archived dependencies to be in this list
	expectedArchivedCount := 0
	for _, status := range repoResults {
		if status.IsArchived {
			expectedArchivedCount++
		}
	}

	if len(archivedDeps) != expectedArchivedCount {
		t.Errorf("Expected %d archived dependencies, got %d", expectedArchivedCount, len(archivedDeps))
	}

	// Verify archived dependency details
	for _, dep := range archivedDeps {
		depMap, ok := dep.(map[string]interface{})
		if !ok {
			t.Fatal("Archived dependency not a map in JSON output")
		}

		path := depMap["module_path"].(string)

		// Check that this dependency is actually archived in our test data
		var found bool
		for _, repo := range repoResults {
			if repo.ModulePath == path && repo.IsArchived {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Dependency %s is in archived list but should not be", path)
		}
	}
}

func TestOutputText(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	moduleInfo := setupTestModuleInfo()
	repoResults := setupRepoStatusResults()

	// Call the function
	OutputText(&moduleInfo, repoResults)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	t.Logf("Text Output: %s", output)

	// Verify the text output contains expected information
	// Use more generic patterns that are likely to be in the output
	expectedPatterns := []string{
		"archived",                 // Look for the word 'archived' instead of a specific heading
		"github.com/archived/repo", // This specific repo should be mentioned
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(strings.ToLower(output), strings.ToLower(pattern)) {
			t.Errorf("Expected output to contain '%s', but it doesn't.", pattern)
		}
	}

	// Make sure there's at least one archived dependency mentioned
	archivedCount := 0
	for _, status := range repoResults {
		if status.IsArchived && strings.Contains(output, status.ModulePath) {
			archivedCount++
		}
	}

	if archivedCount == 0 {
		t.Errorf("Expected at least one archived dependency to be listed in the output")
	}

	// Check that the active repo is not flagged as archived in the output
	for _, status := range repoResults {
		if !status.IsArchived {
			activeRepoArchivedMentioned := strings.Contains(
				strings.ToLower(output),
				strings.ToLower(status.ModulePath+" is archived"))

			if activeRepoArchivedMentioned {
				t.Errorf("Active repo %s should not be indicated as archived", status.ModulePath)
			}
		}
	}
}
