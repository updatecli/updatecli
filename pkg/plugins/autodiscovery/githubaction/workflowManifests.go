package githubaction

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

const (
	// ACTIONKINDDEFAULT is the default kind of action
	// such as "actions/checkout@v2"
	ACTIONKINDDEFAULT = "default"
	// ACTIONKINDLOCAL is the kind of action that is a local path action
	// such as "./actions/checkout"
	ACTIONKINDLOCAL = "local"
	// ACTIONKINDDOCKER is the kind of action that is a docker image
	// such as "docker://alpine:latest"
	ACTIONKINDDOCKER = "docker"
)

// discoverWorkflowManifests discovers all information that could be updated within GitHub action workflow manifests
// then returns a list of Updatecli manifests
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

			if job.Container.Image != "" {
				s := dockerGHAManifestSpec{
					ActionName:        job.Container.Image,
					RelativeFoundFile: relateFoundFile,
					Image:             job.Container.Image,
					TargetKey:         fmt.Sprintf(`$.jobs.%s.container.image`, jobID),
				}

				manifest, err := g.getDockerManifest(&s)
				if err != nil {
					logrus.Errorf("getting GitHub Action manifest: %s", err)
					continue
				}

				if manifest != nil {
					manifests = append(manifests, manifest)
				}
			}

			for stepID, step := range job.Steps {

				if step.Uses == "" {
					// No action to parse
					continue
				}

				URL, owner, repository, directory, reference, actionKind := parseActionName(step.Uses)

				switch actionKind {
				case "":
					logrus.Debugf("GitHub action %q not supported, skipping", step.Uses)
				case ACTIONKINDLOCAL:
					logrus.Debugf("Relative path action %q found, skipping", step.Uses)
					continue
				case ACTIONKINDDOCKER:
					s := dockerGHAManifestSpec{
						ActionName:        step.Uses,
						RelativeFoundFile: relateFoundFile,
						Image:             step.Uses,
						TargetKey:         fmt.Sprintf(`$.jobs.%s.steps[%d].uses`, jobID, stepID),
					}

					manifest, err := g.getDockerManifest(&s)
					if err != nil {
						logrus.Errorf("getting GitHub Action manifest: %s", err)
						continue
					}

					if manifest != nil {
						manifests = append(manifests, manifest)
					}

				case ACTIONKINDDEFAULT:
					u := githubActionManifestSpec{
						URL:               URL,
						Owner:             owner,
						Repository:        repository,
						Directory:         directory,
						Reference:         reference,
						RelativeFoundFile: relateFoundFile,
						CommentDigest:     step.CommentDigest,
						JobID:             jobID,
						StepID:            stepID,
					}

					manifest, err := g.getGitHubActionManifest(&u)
					if err != nil {
						logrus.Errorf("getting GitHub Action manifest: %s", err)
						continue
					}

					if manifest != nil {
						manifests = append(manifests, manifest)
					}
				}
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
