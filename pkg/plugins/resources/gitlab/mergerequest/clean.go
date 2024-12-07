package mergerequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if existing action requires some operations
func (g *Gitlab) CleanAction(report *reports.Action) error {
	logrus.Debugln("cleaning GitLab merge request is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
