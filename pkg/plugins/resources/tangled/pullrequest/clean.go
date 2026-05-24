package pullrequest

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if an existing action requires some operations.
func (t *Tangled) CleanAction(_ context.Context, _ *reports.Action) error {
	logrus.Debugln("cleaning Tangled pull-request is not yet supported")
	return nil
}
