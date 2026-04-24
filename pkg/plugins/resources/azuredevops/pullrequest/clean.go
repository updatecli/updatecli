package pullrequest

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if existing action requires some operations.
func (a *AzureDevOps) CleanAction(ctx context.Context, report *reports.Action) error {
	logrus.Debugln("cleaning Azure DevOps pull requests is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
