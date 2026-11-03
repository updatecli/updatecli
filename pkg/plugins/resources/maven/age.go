package maven

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenversion"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// maxAgeChecks bounds the number of versions dated to find one matching the age filter,
// as every version costs a request to the repository.
const maxAgeChecks = 20

// pomURL returns the url of the pom file of a version, next to the maven-metadata.xml file.
// Credentials in the metadata url are kept, so the request is authenticated the same way.
func (m *Maven) pomURL(metadataURL, ver string) (string, error) {
	u, err := url.Parse(metadataURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(path.Dir(u.Path), ver, fmt.Sprintf("%s-%s.pom", m.spec.ArtifactID, ver))

	return u.String(), nil
}

// releaseDate returns the Last-Modified date of the pom file of a version, which is when it was deployed.
// It returns false when the repository doesn't report one, such as for a SNAPSHOT version whose
// pom file name holds a timestamp.
func (m *Maven) releaseDate(ctx context.Context, metadataURL, ver string) (time.Time, bool, error) {
	pom, err := m.pomURL(metadataURL, ver)
	if err != nil {
		return time.Time{}, false, err
	}

	resp, err := m.request(ctx, http.MethodHead, pom)
	if err != nil {
		return time.Time{}, false, err
	}

	// Some repositories don't accept HEAD requests, or deny them while serving GET requests
	switch resp.code {
	case http.StatusMethodNotAllowed, http.StatusNotImplemented, http.StatusUnauthorized, http.StatusForbidden:
		resp, err = m.request(ctx, http.MethodGet, pom)
		if err != nil {
			return time.Time{}, false, err
		}
	}

	switch {
	case resp.code == http.StatusNotFound || resp.code == http.StatusGone:
		// The pom file name differs, such as for a SNAPSHOT version holding a timestamp
		log.Debugf("no release date for version %q, %s returned %s", ver, redact.URL(pom), resp.status)
		return time.Time{}, false, nil
	case resp.code < 200 || resp.code > 299:
		// Any other failure, such as wrong credentials, mustn't be mistaken for a missing date
		return time.Time{}, false, fmt.Errorf("retrieving the release date of version %q: %s returned %s", ver, redact.URL(pom), resp.status)
	}

	date, err := http.ParseTime(resp.header.Get("Last-Modified"))
	if err != nil {
		log.Debugf("no release date for version %q, %s returned no valid Last-Modified header", ver, redact.URL(pom))
		return time.Time{}, false, nil
	}

	return date, true, nil
}

// headResult holds the status and headers of a response whose body was discarded.
type headResult struct {
	code   int
	status string
	header http.Header
}

// request sends a request and discards its body, only its status and headers are returned.
func (m *Maven) request(ctx context.Context, method, rawURL string) (headResult, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return headResult{}, err
	}

	resp, err := m.webClient.Do(req)
	if err != nil {
		// The http client already masks the password of the url in its errors
		return headResult{}, fmt.Errorf("requesting %s: %w", redact.URL(rawURL), err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	return headResult{code: resp.StatusCode, status: resp.Status, header: resp.Header}, nil
}

// repositoryURL returns the redacted url of the repository holding a maven-metadata.xml file,
// so that messages name the repository to act on.
func repositoryURL(metadataURL string) string {
	return redact.URL(strings.TrimSuffix(metadataURL, "maven-metadata.xml"))
}

// describeRepositories returns "the repository <url>" or "the repositories <url>, <url>", to name
// in messages the repositories a user has to act on.
func describeRepositories(urls []string) string {
	if len(urls) == 1 {
		return "the repository " + urls[0]
	}
	return "the repositories " + strings.Join(urls, ", ")
}

// errNoReleaseDate reports a repository dating none of the versions checked. The source is skipped
// rather than failed, as when the age filter discards every version, so that an age filter set by
// default doesn't break a pipeline that worked without it.
var errNoReleaseDate = errors.New("no release date")

// errNoVersionMatchingAge reports versions the age filter discards, which callers skip rather than fail.
func (m *Maven) errNoVersionMatchingAge() error {
	return fmt.Errorf("%w for the Maven artifact %s/%s", age.ErrNoVersionMatchingAge, m.spec.GroupID, m.spec.ArtifactID)
}

// versionMatchingAge returns the version of a repository selected by the version filter among
// the ones matching the age filter, as the npm, pypi and go module resources do.
func (m *Maven) versionMatchingAge(ctx context.Context, handler mavenmetadata.Handler) (string, error) {
	metadataURL := handler.GetMetadataURL()

	candidates, err := m.ageCandidates(handler)
	if err != nil {
		return "", err
	}

	// discarded is set once the age filter discards a dated version.
	discarded := false

	for _, candidate := range candidates {
		date, dated, err := m.releaseDate(ctx, metadataURL, candidate)
		if err != nil {
			return "", err
		}

		switch {
		case dated && m.spec.Age.Matches(date):
			return candidate, nil
		case dated:
			log.Debugf("ignoring version %q released on %s, outside the age filter", candidate, date)
			discarded = true
		}
	}

	if discarded {
		return "", m.errNoVersionMatchingAge()
	}

	return "", fmt.Errorf("%w: none of the %d versions checked of the Maven artifact %s/%s has one on the repository %s, so the age filter can't be applied: remove it for this repository",
		errNoReleaseDate, len(candidates), m.spec.GroupID, m.spec.ArtifactID, repositoryURL(metadataURL))
}

// ageCandidates returns the versions to check against the age filter, in order of preference,
// at most maxAgeChecks as each one costs a request.
//
// With the kind "latest", they are the version the repository reports as latest, then the versions
// below it, see olderVersions. With other kinds, they are the versions the version filter accepts,
// highest first.
func (m *Maven) ageCandidates(handler mavenmetadata.Handler) ([]string, error) {
	versions, err := handler.GetVersions()
	if err != nil {
		return nil, err
	}

	filter := m.spec.VersionFilter
	if filter.IsZero() || filter.Kind == version.LATESTVERSIONKIND {
		// GetLatestVersion fetches maven-metadata.xml again, but the http cache enabled
		// during a run serves it.
		latest, err := handler.GetLatestVersion()
		if err != nil {
			return nil, err
		}

		return append([]string{latest}, olderVersions(versions, latest, maxAgeChecks-1)...), nil
	}

	var candidates []string
	for len(candidates) < maxAgeChecks {
		found, err := filter.Search(versions)
		if err != nil {
			// A version filter matching no version at all is an error, running out of candidates isn't
			if len(candidates) == 0 {
				return nil, err
			}
			break
		}

		candidate := found.GetVersion()
		candidates = append(candidates, candidate)
		versions = slices.DeleteFunc(versions, func(v string) bool { return v == candidate })
	}

	return candidates, nil
}

// olderVersions returns at most limit versions lower than latest by Maven ordering, highest first.
// Releases come before pre-releases, as the pypi resource does, unless latest is a pre-release.
// Versions are ordered rather than taken in deploy order, so that a backport deployed after a
// newer release, such as 2.12.4 after 2.17.0, isn't preferred to it.
func olderVersions(versions []string, latest string, limit int) []string {
	var older []string
	for _, v := range versions {
		if result, err := mavenversion.Compare(v, latest); err == nil && result < 0 {
			older = append(older, v)
		}
	}

	preferPrerelease := mavenversion.IsPrerelease(latest)
	slices.SortStableFunc(older, func(a, b string) int {
		// Releases first, unless latest is itself a pre-release
		if pa, pb := mavenversion.IsPrerelease(a), mavenversion.IsPrerelease(b); pa != pb && !preferPrerelease {
			if pa {
				return 1
			}
			return -1
		}
		result, _ := mavenversion.Compare(b, a)
		return result
	})

	return older[:min(limit, len(older))]
}
