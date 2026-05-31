package pullrequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/sirupsen/logrus"
	"tangled.org/core/api/tangled"

	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens a Pull Request on the Tangled appview by writing a
// sh.tangled.repo.pull record on the author's PDS. When an open PR already
// exists for the same source/target tuple, a new rounds[] entry is appended
// so the PR reflects the latest patch — Tangled does not auto-refresh the
// review payload when the source branch is updated.
func (t *Tangled) CreateAction(ctx context.Context, report *reports.Action, _ bool) error {
	title := report.Title
	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())
	if err != nil {
		logrus.Warningf("something wrong happened while generating tangled pullrequest body: %s", err)
	}

	if t.spec.Body != "" {
		body = t.spec.Body
	}
	if t.spec.Title != "" {
		title = t.spec.Title
	}

	targetRepoDid, err := t.resolveTargetRepoDID(ctx)
	if err != nil {
		return fmt.Errorf("resolve target repo: %w", err)
	}
	sourceRepoDid := t.resolveSourceRepoDID(targetRepoDid)

	existing, err := t.findExistingPull(ctx, targetRepoDid, sourceRepoDid)
	if err != nil {
		return err
	}

	patch, err := t.generateFormatPatch()
	if err != nil {
		return err
	}
	if strings.TrimSpace(patch) == "" {
		logrus.Infoln("No commits between target and source branches, skipping pullrequest creation")
		return nil
	}

	patchGz, err := gzipPatch(patch)
	if err != nil {
		return fmt.Errorf("gzip patch: %w", err)
	}

	blob, err := t.client.UploadBlob(ctx, patchGz, "application/gzip")
	if err != nil {
		return fmt.Errorf("upload patch blob: %w", err)
	}

	now := time.Now().UTC()
	round := &tangled.RepoPull_Round{
		CreatedAt: now.Format(time.RFC3339),
		PatchBlob: blob,
	}

	if existing != nil {
		record := existing.Record
		record.LexiconTypeID = tangled.RepoPullNSID
		record.Rounds = append(record.Rounds, round)

		if err := t.client.PutRecord(ctx, tangled.RepoPullNSID, existing.Rkey, &record); err != nil {
			return fmt.Errorf("append round to pull request record: %w", err)
		}

		report.Title = record.Title
		if record.Body != nil {
			report.Description = *record.Body
		}
		report.Link = existing.Link
		logrus.Infof("Tangled pullrequest refreshed with new round on %q (rkey=%s, rounds=%d)", existing.Link, existing.Rkey, len(record.Rounds))
		return nil
	}

	source := &tangled.RepoPull_Source{Branch: t.SourceBranch}
	if sourceRepoDid != "" && sourceRepoDid != targetRepoDid {
		s := sourceRepoDid
		source.Repo = &s
	}

	rkey := syntax.NewTIDNow(0).String()
	record := &tangled.RepoPull{
		LexiconTypeID: tangled.RepoPullNSID,
		Title:         title,
		CreatedAt:     now.Format(time.RFC3339),
		Rounds:        []*tangled.RepoPull_Round{round},
		Target: &tangled.RepoPull_Target{
			Repo:   targetRepoDid,
			Branch: t.TargetBranch,
		},
		Source: source,
	}
	if body != "" {
		record.Body = &body
	}

	if err := t.client.PutRecord(ctx, tangled.RepoPullNSID, rkey, record); err != nil {
		return fmt.Errorf("create pull request record: %w", err)
	}

	link := fmt.Sprintf("%s/%s/%s/pulls", t.client.Appview(), t.Owner, t.Repository)

	report.Title = title
	report.Description = body
	report.Link = link

	logrus.Infof("Tangled pullrequest successfully opened on %q (rkey=%s)", link, rkey)
	return nil
}
