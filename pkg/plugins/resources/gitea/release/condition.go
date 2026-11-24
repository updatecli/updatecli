package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	giteaChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/gitea"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (g *Gitea) Condition(ctx context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {

	if scm != nil {
		log.Warningf("Condition not supported for the plugin Gitea Release")
	}

	// A condition checks whether a specific release exists, so the age filter,
	// which only narrows down which release to pick, doesn't apply here.
	releases, err := g.SearchReleases(ctx, age.Spec{})
	if err != nil {
		var unavailable *giteaChangelog.ReleaseListUnavailableError
		if errors.As(err, &unavailable) && g.releasesDisabled(ctx) {
			return false, "releases are disabled for this repository", nil
		}
		return false, "", fmt.Errorf("looking for Gitea release: %w", err)
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}

	if len(releases) == 0 {
		return false, "no Gitea release found", nil
	}

	for _, r := range releases {
		if r == release {
			return true, fmt.Sprintf("Gitea release %q found", release), nil
		}
	}

	return false, fmt.Sprintf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern), nil
}

// releasesDisabled reports only an explicit has_releases: false.
// If the setting can't be read, Condition keeps the original release-list error.
func (g *Gitea) releasesDisabled(ctx context.Context) bool {
	repoURL := g.client.BaseURL.JoinPath("api/v1/repos", url.PathEscape(g.spec.Owner), url.PathEscape(g.spec.Repository))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoURL.String(), nil)
	if err != nil {
		return false
	}

	client := http.Client{
		Transport: g.client.Client.Transport,
		Timeout:   30 * time.Second,
		// Follow repository renames without sending the token to another forge.
		CheckRedirect: func(next *http.Request, via []*http.Request) error {
			if next.URL.Scheme != repoURL.Scheme || next.URL.Host != repoURL.Host || len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	data, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, 1<<20))
	if err != nil {
		return false
	}
	var repo struct {
		HasReleases *bool `json:"has_releases"`
	}
	if err := json.Unmarshal(data, &repo); err != nil || repo.HasReleases == nil {
		return false
	}
	return !*repo.HasReleases
}
