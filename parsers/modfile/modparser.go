package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// ModuleInfo contains relevant information from a go.mod file
type ModuleInfo struct {
	ModuleName string
	GoVersion  string
	Requires   []Dependency
}

// Dependency represents a module dependency
type Dependency struct {
	Path     string
	Version  string
	Indirect bool
}

// ParseGoMod reads and parses a go.mod file from the specified project path
func ParseGoMod(projectPath string) (*ModuleInfo, error) {
	// Find and read the go.mod file
	modPath := filepath.Join(projectPath, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %v", err)
	}

	// Parse the file
	f, err := modfile.Parse(modPath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %v", err)
	}

	// Extract information
	info := &ModuleInfo{
		ModuleName: f.Module.Mod.Path,
		GoVersion:  f.Go.Version,
	}

	// Add dependencies
	for _, req := range f.Require {
		info.Requires = append(info.Requires, Dependency{
			Path:     req.Mod.Path,
			Version:  req.Mod.Version,
			Indirect: req.Indirect,
		})
	}

	return info, nil
}
