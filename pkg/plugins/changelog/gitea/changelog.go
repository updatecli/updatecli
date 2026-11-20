package gitea

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Changelog retrieves release notes from a Gitea compatible forge, such as
// Gitea, Forgejo or Codeberg.
type Changelog struct {
	// URL is the forge base URL, for example https://codeberg.org
	URL string
	// Owner is the repository owner
	Owner string
	// Repository is the repository name
	Repository string
	// Token authenticates API requests. Leave it empty for public repositories
	// on a forge that is not the user's own, so their token is never sent there.
	Token string
	// VersionFilter decides how releases are ordered between from and to
	VersionFilter version.Filter
}

// Search returns release notes between two versions, including both endpoints,
// newest first. For version kinds without range support, it returns only the
// release "to".
func (c *Changelog) Search(from, to string) (result.Changelogs, error) {
	releases, err := publishedReleases(c.catalogID(), to, func() ([]*scm.Release, error) {
		scmClient, err := client.New(client.Spec{URL: c.URL, Token: c.Token})
		if err != nil {
			return nil, fmt.Errorf("creating Gitea client: %w", err)
		}
		return ListReleases(context.Background(), scmClient, c.Owner, c.Repository)
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving Gitea releases: %w", err)
	}

	releases = selectReleases(slices.Clone(releases), c.VersionFilter, from, to)
	if len(releases) == 0 {
		return nil, nil
	}

	baseURL, err := url.Parse(c.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing Gitea URL: %w", err)
	}
	baseURL.User = nil
	releaseURLPrefix := fmt.Sprintf("%s/%s/%s/releases/tag/",
		strings.TrimRight(baseURL.String(), "/"),
		url.PathEscape(c.Owner),
		url.PathEscape(c.Repository),
	)

	changelogs := make(result.Changelogs, 0, len(releases))
	for _, release := range releases {
		link := release.Link
		if link == "" {
			// go-scm may return an empty link even when html_url is present.
			link = releaseURLPrefix + url.PathEscape(release.Tag)
		}
		changelogs = append(changelogs, result.Changelog{
			Title:       release.Tag,
			Body:        release.Description,
			PublishedAt: ReleaseDate(release).String(),
			URL:         link,
		})
	}

	return changelogs, nil
}

// catalogID identifies a repository in the release cache. The token is part
// of the key, hashed, so releases fetched with a user's credentials are never
// served to an anonymous lookup or to another token.
func (c *Changelog) catalogID() string {
	baseURL := strings.TrimRight(c.URL, "/")
	if u, err := url.Parse(baseURL); err == nil {
		u.Host = strings.ToLower(u.Host)
		baseURL = u.String()
	}

	credential := "anonymous"
	if c.Token != "" {
		sum := sha256.Sum256([]byte(c.Token))
		credential = hex.EncodeToString(sum[:])
	}

	return strings.Join([]string{baseURL, c.Owner, c.Repository, credential}, "|")
}

// ReleaseDate returns the date at which a release became public.
// Gitea leaves the publication date empty for releases created without one, so the
// creation date is used as a fallback.
func ReleaseDate(r *scm.Release) time.Time {
	if !r.Published.IsZero() {
		return r.Published
	}
	return r.Created
}
