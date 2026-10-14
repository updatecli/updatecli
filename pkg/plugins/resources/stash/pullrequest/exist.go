package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing Stash pullrequest is already opened.
func (s *Stash) CheckActionExist(ctx context.Context, report *reports.Action) error {

	pullrequestTitle, pullrequestDescription, pullrequestLink, err := s.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		log.Debugf("Stash pull request detected")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	return nil
}
