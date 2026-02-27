package flux

import (
	"bytes"
	"net/url"
	"os"
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

		// Read file content once
		fileData, err := os.ReadFile(foundFluxFile)
		if err != nil {
			logrus.Debugf("Failed reading file %s: %s", foundFluxFile, err)
			continue
		}

		// Split YAML documents
		docs := bytes.Split(fileData, []byte("---"))

		// Process each document separately
		for _, doc := range docs {
			// Skip empty documents
			if len(bytes.TrimSpace(doc)) == 0 {
				continue
			}

			// Retrieve chart dependencies for each chart
			data, err := loadHelmReleaseFromBytes(doc)
			if err != nil {
				logrus.Debugf("Failed loading document from %s as HelmRelease: %s", foundFluxFile, err)
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

			// Skip pipeline if at least one of the helm chart or helm repository is not specified
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

			token := ""
			repoURL, err := url.Parse(helmRepositoryURL)
			switch err {
			case nil:
				if _, ok := f.spec.Auths[repoURL.Host]; ok {
					token = f.spec.Auths[repoURL.Host].Token
					logrus.Debugf("found token for repository %q", repoURL.Host)
				}
			default:
				logrus.Debugf("Ignoring auth configuration due to invalid Helm repository URL: %s", err)
			}

			sourceVersionFilterKind := defaultVersionFilterKind
			sourceVersionFilterPattern := defaultVersionFilterPattern
			sourceVersionFilterRegex := defaultVersionFilterRegex

			if !f.spec.VersionFilter.IsZero() {
				sourceVersionFilterKind = f.versionFilter.Kind
				sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(helmChartVersion)
				sourceVersionFilterRegex = f.versionFilter.Regex
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
				SourceVersionFilterRegex   string
				ScmID                      string
				Token                      string
			}{
				ActionID:                   f.actionID,
				ChartName:                  helmChartName,
				ChartRepository:            helmRepositoryURL,
				File:                       relateFoundFluxFile,
				SourceVersionFilterKind:    sourceVersionFilterKind,
				SourceVersionFilterPattern: sourceVersionFilterPattern,
				SourceVersionFilterRegex:   sourceVersionFilterRegex,
				ScmID:                      f.scmID,
				Token:                      token,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Debugln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	return manifests
}
