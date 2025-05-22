package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseGoMod(t *testing.T) {
	// Setup test directory and files
	testDir := t.TempDir()

	// Test case 1: Valid go.mod file
	t.Run("ValidGoModFile", func(t *testing.T) {
		// Create a valid go.mod file
		validModContent := `module github.com/example/project

go 1.17

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0 // indirect
)
`
		modPath := filepath.Join(testDir, "go.mod")
		err := os.WriteFile(modPath, []byte(validModContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test go.mod: %v", err)
		}

		// Run the parser
		info, err := ParseGoMod(testDir)

		// Verify results
		if err != nil {
			t.Fatalf("ParseGoMod returned error: %v", err)
		}

		// Check module info
		expectedInfo := &ModuleInfo{
			ModuleName: "github.com/example/project",
			GoVersion:  "1.17",
			Requires: []Dependency{
				{
					Path:     "github.com/pkg/errors",
					Version:  "v0.9.1",
					Indirect: false,
				},
				{
					Path:     "github.com/stretchr/testify",
					Version:  "v1.7.0",
					Indirect: true,
				},
			},
		}

		if info.ModuleName != expectedInfo.ModuleName {
			t.Errorf("ModuleName = %q, want %q", info.ModuleName, expectedInfo.ModuleName)
		}

		if info.GoVersion != expectedInfo.GoVersion {
			t.Errorf("GoVersion = %q, want %q", info.GoVersion, expectedInfo.GoVersion)
		}

		if len(info.Requires) != len(expectedInfo.Requires) {
			t.Fatalf("Got %d dependencies, want %d", len(info.Requires), len(expectedInfo.Requires))
		}

		for i, dep := range info.Requires {
			expectedDep := expectedInfo.Requires[i]
			if dep.Path != expectedDep.Path || dep.Version != expectedDep.Version || dep.Indirect != expectedDep.Indirect {
				t.Errorf("Dependency %d mismatch:\ngot: %+v\nwant: %+v", i, dep, expectedDep)
			}
		}
	})

	// Test case 2: go.mod file doesn't exist
	t.Run("NonExistentGoModFile", func(t *testing.T) {
		nonExistentDir := filepath.Join(testDir, "nonexistent")
		err := os.Mkdir(nonExistentDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		_, err = ParseGoMod(nonExistentDir)
		if err == nil {
			t.Fatal("Expected error for missing go.mod file, got nil")
		}
	})

	// Test case 3: Invalid go.mod content
	t.Run("InvalidGoModContent", func(t *testing.T) {
		invalidDir := filepath.Join(testDir, "invalid")
		err := os.Mkdir(invalidDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		invalidModContent := `This is not a valid go.mod file`
		modPath := filepath.Join(invalidDir, "go.mod")
		err = os.WriteFile(modPath, []byte(invalidModContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid go.mod: %v", err)
		}

		_, err = ParseGoMod(invalidDir)
		if err == nil {
			t.Fatal("Expected error for invalid go.mod content, got nil")
		}
	})

	// Test case 4: Empty but valid go.mod file (no dependencies)
	t.Run("EmptyValidGoMod", func(t *testing.T) {
		emptyDir := filepath.Join(testDir, "empty")
		err := os.Mkdir(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		emptyModContent := `module github.com/example/empty
go 1.18
`
		modPath := filepath.Join(emptyDir, "go.mod")
		err = os.WriteFile(modPath, []byte(emptyModContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty go.mod: %v", err)
		}

		info, err := ParseGoMod(emptyDir)
		if err != nil {
			t.Fatalf("ParseGoMod returned error for empty go.mod: %v", err)
		}

		if info.ModuleName != "github.com/example/empty" {
			t.Errorf("ModuleName = %q, want %q", info.ModuleName, "github.com/example/empty")
		}

		if info.GoVersion != "1.18" {
			t.Errorf("GoVersion = %q, want %q", info.GoVersion, "1.18")
		}

		if len(info.Requires) != 0 {
			t.Errorf("Expected 0 dependencies, got %d", len(info.Requires))
		}
	})

	// Test case 5: go.mod with multiple dependencies of mixed types
	t.Run("MixedDependencies", func(t *testing.T) {
		mixedDir := filepath.Join(testDir, "mixed")
		err := os.Mkdir(mixedDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		mixedModContent := `module github.com/example/mixed
go 1.19

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.5.0 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
)
`
		modPath := filepath.Join(mixedDir, "go.mod")
		err = os.WriteFile(modPath, []byte(mixedModContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create mixed go.mod: %v", err)
		}

		info, err := ParseGoMod(mixedDir)
		if err != nil {
			t.Fatalf("ParseGoMod returned error: %v", err)
		}

		expectedDeps := []Dependency{
			{Path: "github.com/pkg/errors", Version: "v0.9.1", Indirect: false},
			{Path: "github.com/sirupsen/logrus", Version: "v1.9.0", Indirect: false},
			{Path: "github.com/spf13/cobra", Version: "v1.5.0", Indirect: true},
			{Path: "github.com/stretchr/testify", Version: "v1.8.0", Indirect: true},
			{Path: "golang.org/x/sys", Version: "v0.0.0-20220811171246-fbc7d0a398ab", Indirect: true},
		}

		if !reflect.DeepEqual(info.Requires, expectedDeps) {
			t.Errorf("Dependencies mismatch.\nGot: %+v\nWant: %+v", info.Requires, expectedDeps)
		}
	})
}
