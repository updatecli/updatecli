package flux

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

func (f Flux) discoverHelmreleaseManifests() [][]byte {

	var manifests [][]byte

	for _, foundFluxFile := range f.helmReleaseFiles {
		logrus.Debugf("parsing helmrelease file %q", foundFluxFile)

		relateFoundFluxFile, err := filepath.Rel(f.rootDir, foundFluxFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		// Retrieve chart dependencies for each chart

		data, err := loadHelmRelease(foundFluxFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		helmChartName := data.Spec.Chart.Spec.Chart
		helmChartVersion := data.Spec.Chart.Spec.Version

		sourceRef := data.Spec.Chart.Spec.SourceRef
		if sourceRef.Namespace == "" {
			sourceRef.Namespace = data.GetNamespace()
		}

		helmRepositoryURL := ""
		for _, helmRepository := range f.helmRepositories {
			if helmRepository.GetName() == sourceRef.Name && helmRepository.GetNamespace() == sourceRef.Namespace {
				helmRepositoryURL = helmRepository.Spec.URL
				break
			}
		}

		// Skip pipeline if at least of the helm chart or helm repository is not specified
		if len(helmChartName) == 0 || len(helmChartVersion) == 0 || len(helmRepositoryURL) == 0 {
			continue
		}

		// If the helmrelease version is not a valid semver, we skip the pipeline
		_, err = semver.NewVersion(helmChartVersion)
		if err != nil {
			if semver.ErrInvalidSemVer == err {
				logrus.Debugf("Ignoring Helm chart %q from %q, as %q not a valid semver version\n", helmChartName, relateFoundFluxFile, helmChartVersion)
				continue
			}
			logrus.Debugf("parsing Helm chart version %q: %s", helmChartVersion, err)
		}

		if len(f.spec.Ignore) > 0 {
			if f.spec.Ignore.isMatchingRules(f.rootDir, relateFoundFluxFile, helmRepositoryURL, helmChartName, helmChartVersion) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as matching ignore rule(s)\n", helmChartName, relateFoundFluxFile)
				continue
			}
		}

		if len(f.spec.Only) > 0 {
			if !f.spec.Only.isMatchingRules(f.rootDir, relateFoundFluxFile, helmRepositoryURL, helmChartName, helmChartVersion) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as not matching only rule(s)\n", helmChartName, relateFoundFluxFile)
				continue
			}
		}

		sourceVersionFilterKind := defaultVersionFilterKind
		sourceVersionFilterPattern := defaultVersionFilterPattern

		if !f.spec.VersionFilter.IsZero() {
			sourceVersionFilterKind = f.versionFilter.Kind
			sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(helmChartVersion)
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = helmChartVersion
			}
		}

		tmpl, err := template.New("manifest").Parse(helmreleaseManifestTemplate)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		params := struct {
			ActionID                   string
			ChartName                  string
			ChartRepository            string
			File                       string
			ImageName                  string
			SourceVersionFilterKind    string
			SourceVersionFilterPattern string
			ScmID                      string
		}{
			ActionID:                   f.actionID,
			ChartName:                  helmChartName,
			ChartRepository:            helmRepositoryURL,
			File:                       relateFoundFluxFile,
			SourceVersionFilterKind:    sourceVersionFilterKind,
			SourceVersionFilterPattern: sourceVersionFilterPattern,
			ScmID:                      f.scmID,
		}

		manifest := bytes.Buffer{}
		if err := tmpl.Execute(&manifest, params); err != nil {
			logrus.Debugln(err)
			continue
		}

		manifests = append(manifests, manifest.Bytes())
	}

	return manifests
}
