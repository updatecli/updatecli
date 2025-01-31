package precommit

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

func (n Precommit) discoverDependencyManifests() ([][]byte, error) {

	var manifests [][]byte

	searchFromDir := n.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if n.spec.RootDir != "" && !path.IsAbs(n.spec.RootDir) {
		searchFromDir = filepath.Join(n.rootDir, n.spec.RootDir)
	}

	foundFiles, err := searchPrecommitConfigFiles(searchFromDir)

	if err != nil {
		return nil, err
	}

	for _, foundFile := range foundFiles {

		logrus.Debugf("parsing file %q", foundFile)

		relativeFoundFile, err := filepath.Rel(n.rootDir, foundFile)
		if err != nil {
			// Let's try the next pom.xml if one fail
			logrus.Debugln(err)
			continue
		}

		data, err := loadPrecommitData(foundFile)

		if err != nil {
			logrus.Debugln(err)
			continue
		}
		if len(data.Repos) == 0 {
			logrus.Errorf("no precommit hook repo found in %q\n", foundFile)
			continue
		}

		for _, repo := range data.Repos {

			if len(n.spec.Ignore) > 0 {
				if n.spec.Ignore.isMatchingRules(n.rootDir, relativeFoundFile, repo.Repo, repo.Rev) {
					logrus.Debugf("Ignoring Hook Repo %q from %q, as matching ignore rule(s)\n", repo.Repo, relativeFoundFile)
					continue
				}
			}

			if len(n.spec.Only) > 0 {
				if !n.spec.Only.isMatchingRules(n.rootDir, relativeFoundFile, repo.Repo, repo.Rev) {
					logrus.Debugf("Ignoring NPM package %q from %q, as not matching only rule(s)\n", repo.Repo, relativeFoundFile)
					continue
				}
			}

			versionPattern, err := n.versionFilter.GreaterThanPattern(repo.Rev)
			if err != nil {
				logrus.Debugf("skipping file %q due to: %s", relativeFoundFile, err)
				continue
			}

			tmpl, err := template.New("manifest").Parse(manifestTemplate)
			if err != nil {
				logrus.Debugln(err)
				continue
			}

			params := struct {
				ManifestName               string
				SourceScmId                string
				SourceScmUrl               string
				SourceID                   string
				SourceName                 string
				SourceKind                 string
				SourceVersionFilterKind    string
				SourceVersionFilterPattern string
				TargetID                   string
				TargetName                 string
				TargetEngine               string
				TargetKey                  string
				File                       string
				ScmID                      string
			}{
				ManifestName:               fmt.Sprintf("Bump %q repo version", repo.Repo),
				SourceScmId:                repo.Repo,
				SourceScmUrl:               repo.Repo,
				SourceName:                 fmt.Sprintf("Get %q repo version", repo.Repo),
				SourceID:                   "gittag",
				SourceKind:                 "gittag",
				SourceVersionFilterKind:    n.versionFilter.Kind,
				SourceVersionFilterPattern: versionPattern,
				TargetID:                   ".pre-commit-config.yaml",
				TargetName:                 fmt.Sprintf("Bump %q repo version to {{ source %q }}", repo.Repo, "gittag"),
				TargetKey:                  fmt.Sprintf("$.repos[?(@.repo == '%s')].rev", repo.Repo),
				TargetEngine:               yaml.EngineYamlPath,
				File:                       relativeFoundFile,
				ScmID:                      n.scmID,
			}

			manifest := bytes.Buffer{}
			if err := tmpl.Execute(&manifest, params); err != nil {
				logrus.Errorln(err)
				continue
			}

			manifests = append(manifests, manifest.Bytes())
		}
	}

	logrus.Printf("%v manifests identified", len(manifests))

	return manifests, nil
}
