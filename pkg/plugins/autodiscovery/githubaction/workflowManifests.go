package githubaction

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

var (
	// referenceLatest specifies the latest version of an action
	referenceLatest = "latest"
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
				action := strings.Split(step.Uses, "@")

				// If the action doesn't contain a reference then we skip it
				if len(action) < 2 {
					continue
				}

				actionName := action[0]
				reference := action[1]

				// If the action name is incomplete then we skip it
				actionNameArray := strings.Split(actionName, "/")
				if len(actionNameArray) < 2 {
					continue
				}

				owner := strings.Split(actionName, "/")[0]
				repository := strings.Split(actionName, "/")[1]

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

				if reference == referenceLatest {
					logrus.Debugf("Ignoring GitHub Action %q as it uses latest tag\n", actionName)
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

				tmpl, err := template.New("manifest").Parse(workflowManifestTemplate)
				if err != nil {
					fmt.Println(err)

					logrus.Debugln(err)
					continue
				}

				params := struct {
					ActionName           string
					Reference            string
					File                 string
					ImageName            string
					JobID                string
					Owner                string
					Repository           string
					VersionFilterKind    string
					VersionFilterPattern string
					StepID               int
					ScmID                string
					Token                string
				}{
					ActionName:           actionName,
					Reference:            reference,
					File:                 relateFoundFile,
					JobID:                jobID,
					Owner:                owner,
					Repository:           repository,
					VersionFilterKind:    versionFilterKind,
					VersionFilterPattern: versionFilterPattern,
					ScmID:                g.scmID,
					StepID:               stepID,
					Token:                g.token,
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

func detectVersionFilter(reference string) (string, string) {

	if _, err := semver.NewVersion(reference); err == nil {
		return "semver", "*"
	}

	return referenceLatest, referenceLatest
}
