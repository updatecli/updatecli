package terraform

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"

	terraformRegistryAddress "github.com/hashicorp/terraform-registry-address"
	"github.com/sirupsen/logrus"
)

func (t Terraform) discoverTerraformProvidersManifests() ([][]byte, error) {
	var manifests [][]byte

	searchFromDir := t.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if t.spec.RootDir != "" && !path.IsAbs(t.spec.RootDir) {
		searchFromDir = filepath.Join(t.rootDir, t.spec.RootDir)
	}

	foundFiles, err := searchTerraformLockFiles(searchFromDir)
	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(t.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}
		providers, err := getTerraformLockContent(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		for provider, providerVersion := range providers {
			// Test if the ignore rule based on path is respected
			if len(t.spec.Ignore) > 0 {
				if t.spec.Ignore.isMatchingRules(t.rootDir, relativeFoundFile, provider, providerVersion) {
					logrus.Debugf("Ignoring provider %q from file %q, as matching ignore rule(s)\n", provider, relativeFoundFile)
					continue
				}
			}

			// Test if the only rule based on path is respected
			if len(t.spec.Only) > 0 {
				if !t.spec.Only.isMatchingRules(t.rootDir, relativeFoundFile, provider, providerVersion) {
					logrus.Debugf("Ignoring provider %q from %q, as not matching only rule(s)\n", provider, relativeFoundFile)
					continue
				}
			}

			versionPattern, err := t.versionFilter.GreaterThanPattern(providerVersion)
			if err != nil {
				logrus.Debugf("skipping provider %q due to: %s", provider, err)
				continue
			}

			moduleManifest, err := t.getTerraformManifest(
				relativeFoundFile,
				provider,
				versionPattern,
			)
			if err != nil {
				logrus.Debugf("skipping provider %q due to: %s", provider, err)
				continue
			}

			manifests = append(manifests, moduleManifest)

		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}

func (t Terraform) getTerraformManifest(filename, provider, versionFilterPattern string) ([]byte, error) {
	tmpl, err := template.New("manifest").Parse(terraformProviderManifestTemplate)
	if err != nil {
		return nil, err
	}

	p, err := terraformRegistryAddress.ParseProviderSource(provider)
	if err != nil {
		return nil, err
	}

	params := struct {
		ActionID             string
		TerraformLockFile    string
		Platforms            []string
		Provider             string
		ProviderNamespace    string
		ProviderName         string
		VersionFilterKind    string
		VersionFilterPattern string
		VersionFilterRegex   string
		ScmID                string
		TargetName           string
	}{
		ActionID:             t.actionID,
		TerraformLockFile:    filename,
		Platforms:            t.spec.Platforms,
		Provider:             p.ForDisplay(),
		ProviderNamespace:    p.Namespace,
		ProviderName:         p.Type,
		VersionFilterKind:    t.versionFilter.Kind,
		VersionFilterPattern: versionFilterPattern,
		VersionFilterRegex:   t.versionFilter.Regex,
		ScmID:                t.scmID,
		TargetName:           fmt.Sprintf("Bump %s to {{ source \"latestVersion\" }}", p.ForDisplay()),
	}

	manifest := bytes.Buffer{}
	if err := tmpl.Execute(&manifest, params); err != nil {
		logrus.Debugln(err)
		return nil, err
	}
	return manifest.Bytes(), nil
}
