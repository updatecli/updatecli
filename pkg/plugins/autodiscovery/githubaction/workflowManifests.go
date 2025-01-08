package githubaction

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"net/url"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

func (g GitHubAction) discoverWorkflowManifests() [][]byte {

	var manifests [][]byte

	for _, foundFile := range g.workflowFiles {
		logrus.Debugf("parsing GitHub Action workflow file %q", foundFile)

		relateFoundFile, err := filepath.Rel(g.rootDir, foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		data, err := loadGitHubActionWorkflow(foundFile)
		if err != nil {
			logrus.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		for jobID, job := range data.Jobs {
			for stepID, step := range job.Steps {
				URL, owner, repository, directory, reference := parseActionName(step.Uses)
				switch reference {
				case "":
					continue
				case "main", "master":
					continue
				}

				actionName := ""
				if URL == "" {
					actionName = owner + "/" + repository
					URL = defaultGitProviderURL
				} else {
					actionName, err = url.JoinPath(URL, owner, repository)
					if err != nil {
						logrus.Errorf("building URL: %s", err)
					}

				}

				kind, token, err := g.getGitProviderKind(URL)
				if err != nil {
					logrus.Debugf("getting credentials: %s", err)
					continue
				}

				if directory != "" {
					actionName = strings.Join([]string{actionName, directory}, "/")
				}

				if len(g.spec.Ignore) > 0 {
					if g.spec.Ignore.isMatchingRules(g.rootDir, relateFoundFile, actionName, reference) {
						logrus.Debugf("Ignoring GitHub Action %q as matching ignore rule(s)\n", actionName)
						continue
					}
				}

				if len(g.spec.Only) > 0 {
					if !g.spec.Only.isMatchingRules(g.rootDir, relateFoundFile, actionName, reference) {
						logrus.Debugf("Ignoring GitHub Action %q as not matching only rule(s)\n", actionName)
						continue
					}
				}

				if slices.Contains([]string{"latest", "master", "main"}, reference) {
					logrus.Debugf("Ignoring GitHub Action %q as it uses the reference %q \n",
						actionName,
						reference,
					)
					continue
				}

				versionFilterKind, versionFilterPattern := detectVersionFilter(reference)
				if !g.spec.VersionFilter.IsZero() {
					versionFilterKind = g.versionFilter.Kind
					versionFilterPattern, err = g.versionFilter.GreaterThanPattern(reference)
					if err != nil {
						logrus.Debugf("building version filter pattern: %s", err)
						versionFilterPattern = reference
					}
				}

				var tmpl *template.Template
				switch kind {
				case kindGitHub:
					tmpl, err = template.New("manifest").Parse(workflowManifestGitHubTemplate)
					if err != nil {
						logrus.Debugln(err)
						continue
					}
				case kindGitea:
					tmpl, err = template.New("manifest").Parse(workflowManifestGiteaTemplate)
					if err != nil {
						logrus.Debugln(err)
						continue
					}
				default:
					logrus.Errorf("unsupported git provider kind %q, skipping", kind)
					continue
				}

				params := struct {
					ActionName           string
					Reference            string
					File                 string
					ImageName            string
					JobID                string
					URL                  string
					Owner                string
					Repository           string
					VersionFilterKind    string
					VersionFilterPattern string
					StepID               int
					ScmID                string
					Token                string
					Digest               bool
				}{
					ActionName:           actionName,
					Reference:            reference,
					File:                 relateFoundFile,
					JobID:                jobID,
					URL:                  URL,
					Owner:                owner,
					Repository:           repository,
					VersionFilterKind:    versionFilterKind,
					VersionFilterPattern: versionFilterPattern,
					ScmID:                g.scmID,
					StepID:               stepID,
					Token:                token,
					Digest:               g.digest,
				}

				manifest := bytes.Buffer{}
				if err := tmpl.Execute(&manifest, params); err != nil {
					logrus.Debugln(err)
					continue
				}

				manifests = append(manifests, manifest.Bytes())
			}
		}

	}

	return manifests
}

// detectVersionFilter tries to identify the kind of versionfilter
func detectVersionFilter(reference string) (string, string) {

	if _, err := semver.NewVersion(reference); err == nil {
		return "semver", "*"
	}

	return "latest", "latest"
}
