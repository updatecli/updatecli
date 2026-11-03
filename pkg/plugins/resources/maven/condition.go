package maven

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
)

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(ctx context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {

	if scm != nil {
		log.Warningf("SCM configuration is not supported for maven condition, aborting")
	}

	if m.spec.Version == "" {
		m.spec.Version = source
	}

	// undated holds the repositories having the version without reporting its release date.
	var undated []string
	// mismatch is set when a repository dates the version outside the age filter, reported when
	// no other repository dates it within, as a mirror may report when it cached the version.
	mismatch := ""
	// errs holds the failures retrieving the release date, reported when no repository dates the
	// version within the age filter, as a failure such as wrong credentials mustn't be hidden.
	var errs []error

	for _, metadataHandler := range m.metadataHandlers {
		// metadataURL contains the URL without username/password
		metadataURL, err := trimUsernamePasswordFromURL(metadataHandler.GetMetadataURL())
		if err != nil {
			return false, "", fmt.Errorf("trying to parse Maven metadata url: %s", err)
		}

		versions, err := metadataHandler.GetVersions()
		if err != nil {
			return false, "", err
		}

		if !slices.Contains(versions, m.spec.Version) {
			continue
		}

		if !m.spec.Age.IsZero() {
			date, dated, err := m.releaseDate(ctx, metadataHandler.GetMetadataURL(), m.spec.Version)
			if err != nil {
				// Another repository may report it, as Source does
				errs = append(errs, err)
				continue
			}
			if !dated {
				// Another repository may report it
				undated = append(undated, repositoryURL(metadataHandler.GetMetadataURL()))
				continue
			}
			if !m.spec.Age.Matches(date) {
				// Another repository may date it within the age filter
				mismatch = fmt.Sprintf("Version %s released on %s does not match the age filter for Maven artifact (%s/%s) on the Maven Repository (%s)",
					m.spec.Version, date, m.spec.GroupID, m.spec.ArtifactID, repositoryURL(metadataHandler.GetMetadataURL()))
				continue
			}
		}

		return true, fmt.Sprintf("Version %s is available on the Maven Repository (%s)",
			m.spec.Version, redact.URL(metadataURL)), nil
	}

	if len(errs) > 0 {
		return false, "", errors.Join(errs...)
	}

	if mismatch != "" {
		return false, mismatch, nil
	}

	if len(undated) > 0 {
		return false, fmt.Sprintf("Version %s of Maven artifact (%s/%s) has no release date on %s, so the age filter can't be applied: remove it",
			m.spec.Version, m.spec.GroupID, m.spec.ArtifactID, describeRepositories(undated)), nil
	}

	return false, fmt.Sprintf("Version %s is not found for Maven artifact (%s/%s)",
		m.spec.Version, m.spec.GroupID, m.spec.ArtifactID), nil
}
