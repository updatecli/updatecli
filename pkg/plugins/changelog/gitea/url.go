package gitea

import (
	"net/url"
	"slices"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// knownHosts lists public forges that serve the Gitea releases API.
	// Only these hosts are recognized from artifact metadata, since an
	// unknown host cannot be told apart from another forge without a request.
	knownHosts = []string{"codeberg.org", "gitea.com"}
)

// IsKnownHost reports whether host is a public Gitea compatible forge.
func IsKnownHost(host string) bool {
	return slices.Contains(knownHosts, strings.ToLower(host))
}

// ParseRepositoryURL extracts the forge base URL, owner and repository from a
// repository reference on a known Gitea compatible host. It accepts
// https://host/owner/repo[.git][/...], git+https://..., git+ssh://git@host/owner/repo.git,
// git@host:owner/repo.git and the bare host/owner/repo form used by Go modules.
// ok is false when the reference cannot be parsed or its host is not known.
func ParseRepositoryURL(raw string) (baseURL, owner, repository string, ok bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "git+")

	var host, path string
	switch {
	case strings.Contains(raw, "://"):
		u, err := url.Parse(raw)
		if err != nil {
			return "", "", "", false
		}
		host, path = u.Hostname(), u.Path
	case strings.HasPrefix(raw, "git@"):
		// scp-like syntax: git@host:owner/repo.git
		var found bool
		host, path, found = strings.Cut(strings.TrimPrefix(raw, "git@"), ":")
		if !found {
			return "", "", "", false
		}
	default:
		host, path, _ = strings.Cut(raw, "/")
	}

	if !IsKnownHost(host) {
		return "", "", "", false
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", false
	}

	repository = strings.TrimSuffix(parts[1], ".git")
	if repository == "" {
		return "", "", "", false
	}

	return "https://" + strings.ToLower(host), parts[0], repository, true
}

// SearchRepository returns release notes between from and to for a repository
// reference on a known Gitea compatible host, such as a package's source URL.
// It returns nil when the host is not known, the lookup fails or no release
// matches. The request is anonymous, so no user token is sent to a forge
// the user did not configure.
func SearchRepository(rawURL, from, to string) *result.Changelogs {
	baseURL, owner, repository, ok := ParseRepositoryURL(rawURL)
	if !ok {
		return nil
	}

	changelog := Changelog{
		URL:        baseURL,
		Owner:      owner,
		Repository: repository,
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		log.Debugf("searching Gitea releases for %s/%s on %s: %s", owner, repository, baseURL, err)
		return nil
	}
	if len(releases) == 0 {
		return nil
	}

	return &releases
}
