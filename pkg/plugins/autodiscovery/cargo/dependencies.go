package cargo

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"text/template"
)

var (
	// CargoValidFiles specifies accepted Cargo files.
	CargoValidFiles [1]string = [1]string{"Cargo.toml"}
)

type crateDependency struct {
	Name     string
	Registry string
	Version  string
}

type crateMetadata struct {
	Name            string
	Dependencies    []crateDependency
	DevDependencies []crateDependency
}

func generateManifest(crateName string, dependency crateDependency, relativeFile string, foundFile string, dependencyType string) (bytes.Buffer, error) {
	manifest := bytes.Buffer{}
	tmpl, err := template.New("manifest").Parse(dependencyManifest)
	if err != nil {
		logrus.Debugln(err)
		return manifest, err
	}
	params := struct {
		ManifestName               string
		CrateName                  string
		DependencyName             string
		SourceID                   string
		SourceName                 string
		SourceVersionFilterKind    string
		SourceVersionFilterPattern string
		ExistingSourceID           string
		ExistingSourceName         string
		ExistingSourceKey          string
		ConditionID                string
		ConditionQuery             string
		File                       string
		TargetID                   string
		TargetFile                 string
		TargetKey                  string
		ScmID                      string
	}{
		ManifestName:               fmt.Sprintf("Bump %s %q for \"%s\" crate", dependencyType, dependency.Name, crateName),
		CrateName:                  crateName,
		DependencyName:             dependency.Name,
		SourceID:                   dependency.Name,
		SourceName:                 fmt.Sprintf("Get latest %q crate version", dependency.Name),
		SourceVersionFilterKind:    "semver",
		SourceVersionFilterPattern: "*",
		ExistingSourceID:           fmt.Sprintf("%s-current-version", dependency.Name),
		ExistingSourceKey:          fmt.Sprintf("%s.%s.version", dependencyType, dependency.Name),
		ExistingSourceName:         fmt.Sprintf("Get current %q crate version", dependency.Name),
		ConditionID:                dependency.Name,
		ConditionQuery:             fmt.Sprintf("%s.(?:-=%s).version", dependencyType, dependency.Name),
		File:                       relativeFile,
		TargetID:                   dependency.Name,
		TargetFile:                 filepath.Base(foundFile),
		TargetKey:                  fmt.Sprintf("%s.%s.version", dependencyType, dependency.Name),
	}

	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return manifest, err
	}
	return manifest, nil
}

func (c Cargo) discoverCargoDependenciesManifests() ([][]byte, error) {
	manifests := [][]byte{}

	foundCargoFiles, err := findCargoFiles(
		c.rootDir,
		CargoValidFiles[:],
	)

	if err != nil {
		return nil, err
	}

	for _, foundCargoFile := range foundCargoFiles {
		logrus.Debugf("parsing file %q", foundCargoFile)

		relativeFoundCargoFile, err := filepath.Rel(c.rootDir, foundCargoFile)
		if err != nil {
			// Jump to the next Cargo if current failed
			logrus.Debugln(err)
			continue
		}

		cargoRelativePath := filepath.Dir(relativeFoundCargoFile)
		cargoCrateName := filepath.Base(cargoRelativePath)

		// Test if the ignore rule based on path doesn't match
		if len(c.spec.Ignore) > 0 && c.spec.Ignore.isMatchingIgnoreRule(c.rootDir, relativeFoundCargoFile) {
			logrus.Debugf("Ignoring Cargo Crate %q from %q, as not matching rule(s)\n",
				cargoCrateName,
				cargoRelativePath)
			continue
		}

		// Test if the only rule based on path match
		if len(c.spec.Only) > 0 && !c.spec.Only.isMatchingOnlyRule(c.rootDir, relativeFoundCargoFile) {
			logrus.Debugf("Ignoring Cargo Crate %q from %q, as not matching rule(s)\n",
				cargoCrateName,
				cargoRelativePath)
			continue
		}

		// Retrieve Cargo dependencies for each crate
		crate, err := getCrateMetadata(c.rootDir, foundCargoFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if crate == nil {
			continue
		}

		if len(crate.Dependencies) == 0 && len(crate.DevDependencies) == 0 {
			continue
		}

		c := *crate
		for _, dependency := range c.Dependencies {
			manifest, err := generateManifest(c.Name, dependency, relativeFoundCargoFile, foundCargoFile, "dependencies")
			if err != nil {
				logrus.Debugln(err)
				continue
			}
			manifests = append(manifests, manifest.Bytes())
		}
		for _, dependency := range c.DevDependencies {
			manifest, err := generateManifest(c.Name, dependency, relativeFoundCargoFile, foundCargoFile, "dev-dependencies")
			if err != nil {
				logrus.Debugln(err)
				continue
			}
			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests, nil
}
