package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

const (
	// pageSize is the number of releases requested per API page.
	pageSize = 30
	// pageTimeout bounds each API page request.
	pageTimeout = 30 * time.Second
)

// ReleaseListUnavailableError means the first page of releases returned 403 or 404.
// A condition can check whether the repository has the releases unit disabled.
type ReleaseListUnavailableError struct{ err error }

func (e *ReleaseListUnavailableError) Error() string { return e.err.Error() }
func (e *ReleaseListUnavailableError) Unwrap() error { return e.err }

// ListReleases retrieves every release of a repository, newest first,
// without applying any filter. Drafts are included.
func ListReleases(ctx context.Context, c client.Client, owner, repository string) ([]*scm.Release, error) {
	scmClient := (*scm.Client)(c)
	var results []*scm.Release
	page := 1
	for {
		pageCtx, cancel := context.WithTimeout(ctx, pageTimeout)
		releases, resp, err := scmClient.Releases.List(
			pageCtx,
			strings.Join([]string{owner, repository}, "/"),
			scm.ReleaseListOptions{
				Page:   page,
				Size:   pageSize,
				Open:   true,
				Closed: true,
			},
		)
		cancel()
		if err != nil {
			if page == 1 && resp != nil && (resp.Status == http.StatusForbidden || resp.Status == http.StatusNotFound) {
				return nil, &ReleaseListUnavailableError{err}
			}
			return nil, err
		}

		results = append(results, releases...)
		// An empty page means there is nothing left, whatever the Link header says.
		if len(releases) == 0 || resp.Page.Next == 0 {
			return results, nil
		}
		// A next link that does not move forward would make this loop forever.
		if resp.Page.Next <= page {
			return nil, fmt.Errorf("gitea release pagination did not advance: page %d links to next page %d", page, resp.Page.Next)
		}
		page = resp.Page.Next
	}
}
