package helmfile

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
)

var (
	// DefaultFilePattern specifies accepted Helm chart metadata filename
	DefaultFilePattern [2]string = [2]string{"*.yaml", "*.yml"}
)

// Release holds the Helmfile release information.
type release struct {
	Name    string
	Chart   string
	Version string
}

// Repository holds the Helmfile repository information
type repository struct {
	Name     string
	URL      string
	OCI      bool
	Username string
	Password string
}

// helmfileMetadata is the information retrieved from Helmfile files.
type helmfileMetadata struct {
	Name         string
	Repositories []repository
	Releases     []release
}

// discoverHelmfileReleaseManifests search recursively from a root directory for Helmfile file
func (h Helmfile) discoverHelmfileReleaseManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := h.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if h.spec.RootDir != "" && !path.IsAbs(h.spec.RootDir) {
		searchFromDir = filepath.Join(h.rootDir, h.spec.RootDir)
	}

	foundHelmfileFiles, err := searchHelmfileFiles(
		searchFromDir,
		DefaultFilePattern[:])

	if err != nil {
		return nil, err
	}

	for _, foundHelmfile := range foundHelmfileFiles {
		logrus.Debugf("parsing file %q", foundHelmfile)

		relativeFoundChartFile, err := filepath.Rel(h.rootDir, foundHelmfile)
		if err != nil {
			// Jump to the next Helmfile if current failed
			logrus.Debugln(err)
			continue
		}

		helmfileRelativeMetadataPath := filepath.Dir(relativeFoundChartFile)
		helmfileFilename := filepath.Base(helmfileRelativeMetadataPath)

		// Retrieve chart dependencies for each chart

		metadata, err := getHelmfileMetadata(foundHelmfile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if metadata == nil {
			continue
		}

		if len(metadata.Releases) == 0 {
			continue
		}

		for i, release := range metadata.Releases {
			var chartName, chartURL, OCIUsername, OCIPassword string
			var isOCI bool

			for _, repository := range metadata.Repositories {
				if strings.HasPrefix(release.Chart, repository.Name+"/") {
					chartName = strings.TrimPrefix(release.Chart, repository.Name+"/")
					chartURL = repository.URL
					isOCI = repository.OCI
					OCIUsername = repository.Username
					OCIPassword = repository.Password
					break
				}
			}

			if chartName == "" || chartURL == "" {
				logrus.Debugf("repository not identified for release %q, skipping", release.Chart)
				continue
			}

			// Helmfile uses the repository flag 'oci'
			// to identify OCI Helm chart
			// Updatecli expects the scheme 'oci://'.
			// Therefor Updatecli removes any 'http://' or 'https://' schemes before adding 'oci://'
			if isOCI {
				for _, scheme := range []string{"https://", "http://"} {
					if strings.HasPrefix(chartURL, scheme) {
						chartURL = strings.TrimPrefix(chartURL, scheme)
						break
					}
				}
				chartURL = "oci://" + chartURL
			}

			if release.Version == "" {
				logrus.Debugf("no version specified for release %q, skipping", release.Chart)
				continue
			}

			helmSourcespec := helm.Spec{
				Name: chartName,
				URL:  chartURL,
			}
			if OCIUsername != "" && isOCI {
				helmSourcespec.InlineKeyChain.Username = OCIUsername
			}
			if OCIPassword != "" && isOCI {
				helmSourcespec.InlineKeyChain.Password = OCIPassword
			}

			sourceVersionFilterKind := "semver"
			sourceVersionFilterPattern := "*"

			if !h.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = h.versionFilter.Kind
				sourceVersionFilterPattern, err = h.versionFilter.GreaterThanPattern(release.Version)
				if err != nil {
					logrus.Debugf("building version filter pattern: %s", err)
					sourceVersionFilterPattern = "*"
				}
			}

			if len(h.spec.Ignore) > 0 {
				if h.spec.Ignore.isMatchingRules(h.rootDir, relativeFoundChartFile, chartURL, chartName, release.Version) {
					logrus.Debugf("Ignoring Helmfile release %q from %q, as matching ignore rule(s)\n", chartURL, helmfileFilename)
					continue
				}
			}

			if len(h.spec.Only) > 0 {
				if !h.spec.Only.isMatchingRules(h.rootDir, relativeFoundChartFile, chartURL, chartName, release.Version) {
					logrus.Debugf("Ignoring Helmfile release %q from %q, as not matching only rule(s)\n", chartURL, helmfileFilename)
					continue
				}
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			params := struct {
				ActionID                   string
				ManifestName               string
				ChartName                  string
				ChartRepository            string
				ConditionID                string
				ConditionName              string
				ConditionKey               string
				ConditionValue             string
				SourceID                   string
				SourceName                 string
				SourceKind                 string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				TargetID                   string
				TargetName                 string
				TargetKey                  string
				File                       string
				ScmID                      string
			}{
				ActionID:                   h.actionID,
				ManifestName:               fmt.Sprintf("Bump %q Helm Chart version for Helmfile %q", release.Name, relativeFoundChartFile),
				ChartName:                  chartName,
				ChartRepository:            chartURL,
				ConditionID:                release.Name,
				ConditionName:              fmt.Sprintf("Ensure release %q is specified for Helmfile %q", release.Name, relativeFoundChartFile),
				ConditionKey:               fmt.Sprintf("$.releases[%d].chart", i),
				ConditionValue:             release.Chart,
				SourceID:                   release.Name,
				SourceName:                 fmt.Sprintf("Get latest %q Helm Chart version", release.Name),
				SourceKind:                 "helmchart",
				SourceVersionFilterKind:    sourceVersionFilterKind,
				SourceVersionFilterPattern: sourceVersionFilterPattern,
				TargetID:                   release.Name,
				TargetName:                 fmt.Sprintf("deps(helmfile): update %q Helm Chart version to {{ source %q}}", release.Name, release.Name),
				TargetKey:                  fmt.Sprintf("$.releases[%d].version", i),
				File:                       foundHelmfile,
				ScmID:                      h.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())

		}
	}

	return manifests, nil
}
