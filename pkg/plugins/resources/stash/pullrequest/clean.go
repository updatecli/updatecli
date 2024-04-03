package pullrequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verify if existing action requires some operations
func (s *Stash) CleanAction(report reports.Action) error {
	logrus.Debugln("cleaning Gitea pull-request is not yet supported. Feel free to open an issue to mark your interest.")
	return nil
}
