package mergerequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if existing action requires some operations
func (g *Gitlab) CleanAction(ctx context.Context, report *reports.Action) error {
	log.Debugln("cleaning GitLab merge request is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
