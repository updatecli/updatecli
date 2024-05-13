package argocd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
)

var (
	// ArgoCDFilePatterns specifies accepted Helm chart metadata file name
	ArgoCDFilePatterns [2]string = [2]string{"*.yaml", "*.yml"}
)

// ArgoCDApplicationSpec is the information that we need to retrieve from Helm chart files.
type ArgoCDApplicationSpec struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Source struct {
			RepoURL        string `yaml:"repoURL"`
			TargetRevision string `yaml:"targetRevision"`
			Chart          string `yaml:"chart"`
		}
	}
}

func (f ArgoCD) discoverArgoCDManifests() ([][]byte, error) {

	var manifests [][]byte

	foundFiles, err := searchArgoCDFiles(
		f.rootDir,
		ArgoCDFilePatterns[:])

	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {
		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(f.rootDir, foundFile)
		if err != nil {
			// Let's try the next chart if one fail
			logrus.Debugln(err)
			continue
		}

		// Retrieve chart dependencies for each chart

		data, err := loadApplicationData(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		// Skip pipeline if at least of the helm chart or helm repository is not specified
		if len(data.Spec.Source.Chart) == 0 || len(data.Spec.Source.RepoURL) == 0 {
			continue
		}

		if len(f.spec.Ignore) > 0 {
			if f.spec.Ignore.isMatchingRules(f.rootDir, relativeFoundFile, data.Spec.Source.RepoURL, data.Spec.Source.Chart, data.Spec.Source.TargetRevision) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as matching ignore rule(s)\n", data.Spec.Source.Chart, relativeFoundFile)
				continue
			}
		}

		if len(f.spec.Only) > 0 {
			if !f.spec.Only.isMatchingRules(f.rootDir, relativeFoundFile, data.Spec.Source.RepoURL, data.Spec.Source.Chart, data.Spec.Source.TargetRevision) {
				logrus.Debugf("Ignoring Helm chart %q from %q, as not matching only rule(s)\n", data.Spec.Source.Chart, relativeFoundFile)
				continue
			}
		}

		sourceVersionFilterKind := "semver"
		sourceVersionFilterPattern := "*"

		if !f.spec.VersionFilter.IsZero() {
			sourceVersionFilterKind = f.versionFilter.Kind
			sourceVersionFilterPattern, err = f.versionFilter.GreaterThanPattern(data.Spec.Source.TargetRevision)
			if err != nil {
				logrus.Debugf("building version filter pattern: %s", err)
				sourceVersionFilterPattern = "*"
			}
		}

		tmpl, err := template.New("manifest").Parse(manifestTemplate)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		params := struct {
			ManifestName               string
			ImageName                  string
			ChartName                  string
			ChartRepository            string
			ManifestFile               string
			ConditionID                string
			SourceID                   string
			SourceName                 string
			SourceKind                 string
			SourceVersionFilterKind    string
			SourceVersionFilterPattern string
			TargetID                   string
			File                       string
			ScmID                      string
		}{
			ManifestName:               fmt.Sprintf("deps(helm): bump Helm chart %q in ArgoCD manifest %q", data.Spec.Source.Chart, relativeFoundFile),
			ChartName:                  data.Spec.Source.Chart,
			ChartRepository:            data.Spec.Source.RepoURL,
			ConditionID:                data.Spec.Source.Chart,
			ManifestFile:               relativeFoundFile,
			SourceID:                   data.Spec.Source.Chart,
			SourceName:                 fmt.Sprintf("Get latest %q Helm chart version", data.Spec.Source.Chart),
			SourceKind:                 "helmchart",
			SourceVersionFilterKind:    sourceVersionFilterKind,
			SourceVersionFilterPattern: sourceVersionFilterPattern,
			TargetID:                   data.Spec.Source.Chart,
			File:                       relativeFoundFile,
			ScmID:                      f.scmID,
		}

		manifest := bytes.Buffer{}
		if err := tmpl.Execute(&manifest, params); err != nil {
			logrus.Debugln(err)
			continue
		}

		manifests = append(manifests, manifest.Bytes())
	}

	return manifests, nil
}
