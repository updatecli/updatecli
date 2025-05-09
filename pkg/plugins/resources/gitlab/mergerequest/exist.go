package mergerequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing GitLab merge request is already opened.
func (g *Gitlab) CheckActionExist(report *reports.Action) error {

	mr, err := g.findExistingMR()
	if err != nil {
		return err
	}

	if mr != nil {
		logrus.Debugf("GitLab merge request detected")

		report.Title = mr.Title
		report.Link = mr.WebURL
		report.Description = mr.Description
		return nil
	}

	return nil
}
