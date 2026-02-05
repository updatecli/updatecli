package bazel

import (
	"fmt"
	"os"

	"github.com/updatecli/updatecli/pkg/plugins/resources/bazelmod"
)

// dependency represents a single Bazel module dependency
type dependency struct {
	Name    string
	Version string
	File    string // Path to the MODULE.bazel file
}

// parseModuleDependencies extracts all bazel_dep() declarations from a MODULE.bazel file
func parseModuleDependencies(moduleFile string) ([]dependency, error) {
	content, err := os.ReadFile(moduleFile)
	if err != nil {
		return nil, fmt.Errorf("reading MODULE.bazel file %q: %w", moduleFile, err)
	}

	moduleFileParsed, err := bazelmod.ParseModuleFile(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing MODULE.bazel file %q: %w", moduleFile, err)
	}

	dependencies := []dependency{}
	for _, dep := range moduleFileParsed.Deps {
		// Only include dependencies that have both name and version
		if dep.Name != "" && dep.Version != "" {
			dependencies = append(dependencies, dependency{
				Name:    dep.Name,
				Version: dep.Version,
				File:    moduleFile,
			})
		}
	}

	return dependencies, nil
}
