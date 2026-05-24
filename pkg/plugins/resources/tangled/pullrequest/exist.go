package pullrequest

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing Tangled pullrequest is already opened.
func (t *Tangled) CheckActionExist(ctx context.Context, report *reports.Action) error {
	targetRepoDid, err := t.resolveTargetRepoDID(ctx)
	if err != nil {
		return fmt.Errorf("resolve target repo: %w", err)
	}
	sourceRepoDid := t.resolveSourceRepoDID(targetRepoDid)

	existing, err := t.findExistingPull(ctx, targetRepoDid, sourceRepoDid)
	if err != nil {
		return err
	}

	if existing != nil {
		logrus.Debugf("Tangled pull request detected")
		report.Link = existing.Link
	}

	return nil
}
