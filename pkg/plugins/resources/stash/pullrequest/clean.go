package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if an existing action requires some operations
func (s *Stash) CleanAction(ctx context.Context, report *reports.Action) error {
	log.Debugln("cleaning Stash pull-request is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
