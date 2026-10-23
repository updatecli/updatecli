package maven

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Source return the latest version
func (m *Maven) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {

	// errs holds the lookup failure of each repository, reported when none returns a version.
	var errs []error
	// undated holds the repositories dating none of the versions checked against the age filter.
	var undated []string

	for _, metadataHandler := range m.metadataHandlers {
		// metadataURL contains the URL without username/password
		metadataURL, err := trimUsernamePasswordFromURL(metadataHandler.GetMetadataURL())
		if err != nil {
			log.Errorf("Trying to parse Maven metadata url: %s", err)
		}

		latestVersion, err := m.latestVersion(ctx, metadataHandler)
		if err != nil {
			errs = append(errs, err)
			if errors.Is(err, errNoReleaseDate) {
				undated = append(undated, repositoryURL(metadataHandler.GetMetadataURL()))
			}
			// The age filter holding back every version is expected, not worth a warning
			if errors.Is(err, age.ErrNoVersionMatchingAge) {
				log.Debugf("getting latest version: %v", err)
			} else {
				log.Warnf("getting latest version: %v", err)
			}
		}

		if latestVersion != "" {
			resultSource.Result = result.SUCCESS
			resultSource.Information = latestVersion
			description := "Latest version"
			if !m.spec.Age.IsZero() {
				description = "Latest version matching the age filter"
			}
			resultSource.Description = fmt.Sprintf("%s is %s on the Maven repository at %s",
				description, latestVersion, redact.URL(metadataURL))
			return nil
		}

	}

	/*
		The age filter discarding the versions of a repository, or not being able to date them, is
		an expected state rather than a failure, so the source is skipped instead. Any other
		failure, such as wrong credentials, isn't hidden behind a skip.
	*/
	if isAgeHold(errs) {
		resultSource.Result = result.SKIPPED
		resultSource.Description = fmt.Sprintf("no version of the Maven artifact %s/%s matches the age filter yet", m.spec.GroupID, m.spec.ArtifactID)
		if len(undated) > 0 {
			resultSource.Description = fmt.Sprintf("the age filter can't be applied to the Maven artifact %s/%s, as %s reports no release date: remove it",
				m.spec.GroupID, m.spec.ArtifactID, describeRepositories(undated))
		}
		return nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("no latest version for the Maven Artifact %s/%s: %w", m.spec.GroupID, m.spec.ArtifactID, errors.Join(errs...))
	}

	return fmt.Errorf("no latest version for the Maven Artifact %s/%s", m.spec.GroupID, m.spec.ArtifactID)
}

// latestVersion returns the version of a repository selected by the version filter and, when set, the age filter.
func (m *Maven) latestVersion(ctx context.Context, metadataHandler mavenmetadata.Handler) (string, error) {
	if m.spec.Age.IsZero() {
		return metadataHandler.GetLatestVersion()
	}

	return m.versionMatchingAge(ctx, metadataHandler)
}

// isAgeHold returns true when the age filter discarded or couldn't date the versions of at least
// one repository and the other repositories don't host the artifact, as a repository declared in
// a pom.xml or a settings.xml file rarely hosts every artifact.
func isAgeHold(errs []error) bool {
	held := false

	for _, err := range errs {
		switch {
		case errors.Is(err, age.ErrNoVersionMatchingAge), errors.Is(err, errNoReleaseDate):
			held = true
		case errors.Is(err, text.ErrNotFound):
		default:
			return false
		}
	}

	return held
}
