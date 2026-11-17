package githubaction

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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

// discoverManifests discovers all information that could be updated within GitHub action workflow manifests and composite actions
// then returns a list of Updatecli manifests
func (g GitHubAction) discoverManifests() [][]byte {

	var manifests [][]byte

	for _, foundAction := range g.compositeActionFiles {
		log.Debugf("parsing Composite Action file %q", foundAction)

		relateFoundFile, err := filepath.Rel(g.rootDir, foundAction)
		if err != nil {
			log.Debugln(err)
			continue
		}

		action, err := loadCompositeAction(foundAction)
		if err != nil {
			log.Debugln(err)
			continue
		}

		if action == nil {
			continue
		}

		for stepID, step := range action.Job.Steps {

			if step.Uses == "" {
				// No action to parse
				continue
			}

			URL, owner, repository, directory, reference, actionKind := parseActionName(step.Uses)

			switch actionKind {
			case "":
				log.Debugf("GitHub action %q not supported, skipping", step.Uses)
			case ACTIONKINDLOCAL:
				log.Debugf("Relative path action %q found, skipping", step.Uses)
				continue
			case ACTIONKINDDOCKER:
				s := dockerGHAManifestSpec{
					ActionName:        step.Uses,
					RelativeFoundFile: relateFoundFile,
					Image:             step.Uses,
					TargetKey:         fmt.Sprintf(`$.runs.steps[%d].uses`, stepID),
				}

				manifest, err := g.getDockerManifest(&s)
				if err != nil {
					log.Errorf("getting GitHub Action manifest: %s", err)
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
					Composite:         true,
					StepID:            stepID,
				}

				manifest, err := g.getGitHubActionManifest(&u)
				if err != nil {
					log.Errorf("getting GitHub Action manifest: %s", err)
					continue
				}

				if manifest != nil {
					manifests = append(manifests, manifest)
				}
			}
		}

	}
	for _, foundFile := range g.workflowFiles {
		log.Debugf("parsing GitHub Action workflow file %q", foundFile)

		relateFoundFile, err := filepath.Rel(g.rootDir, foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		data, err := loadGitHubActionWorkflow(foundFile)
		if err != nil {
			log.Debugln(err)
			continue
		}

		if data == nil {
			continue
		}

		// A workflow onboarded in the lockfile only runs the actions the lockfile records, so its
		// actions are only bumped when "gh actions-lock" can update the lockfile too.
		// An unreadable lockfile may lock any workflow it sits next to.
		relativeWorkflow := filepath.ToSlash(relateFoundFile)
		lockWorkflow := g.lockedWorkflows[relativeWorkflow]
		skipActions := (lockWorkflow && !g.lockAvailable) ||
			(g.lockfileInvalid && path.Dir(relativeWorkflow) == path.Dir(actionsLockFile))
		switch {
		case lockWorkflow && !g.lockAvailable:
			log.Warningf("skipping the actions of %q: the workflow is locked in %q but \"gh actions-lock\" can't update it",
				relateFoundFile, actionsLockFile)
		case skipActions:
			log.Warningf("skipping the actions of %q: the workflow may be locked in %q, which can't be read",
				relateFoundFile, actionsLockFile)
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
					log.Errorf("getting GitHub Action manifest: %s", err)
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
					log.Debugf("GitHub action %q not supported, skipping", step.Uses)
				case ACTIONKINDLOCAL:
					log.Debugf("Relative path action %q found, skipping", step.Uses)
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
						log.Errorf("getting GitHub Action manifest: %s", err)
						continue
					}

					if manifest != nil {
						manifests = append(manifests, manifest)
					}

				case ACTIONKINDDEFAULT:
					if skipActions {
						continue
					}

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
						LockWorkflow:      lockWorkflow,
					}

					manifest, err := g.getGitHubActionManifest(&u)
					if err != nil {
						log.Errorf("getting GitHub Action manifest: %s", err)
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

// detectVersionFilter tries to identify the kind of versionfilter.
// With an age filter that doesn't allow a downgrade, a semver reference is used as lower bound,
// so that a release filtered out by the age filter never makes the source fall back to an older
// one. A prerelease reference, such as "v2.0.0-rc.2", keeps accepting the newer prereleases.
func detectVersionFilter(reference string, releaseAge age.AutodiscoverySpec) (string, string) {

	if _, err := semver.NewVersion(reference); err == nil {
		switch {
		case releaseAge.NeverDowngrade():
			return version.SEMVERVERSIONKIND, version.BoundPatternAcceptingPrereleases("*", reference)
		case releaseAge.MayDowngrade():
			return version.SEMVERVERSIONKIND, version.UnboundedPattern(reference)
		default:
			return version.SEMVERVERSIONKIND, "*"
		}
	}

	return latestVersionIdentifier, latestVersionIdentifier
}
