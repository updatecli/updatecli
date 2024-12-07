package pullrequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if an existing action requires some operations
func (b *Bitbucket) CleanAction(report *reports.Action) error {
	logrus.Debugln("cleaning Bitbucket Cloud pull request is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
