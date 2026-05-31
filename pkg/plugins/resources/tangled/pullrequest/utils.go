package pullrequest

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"tangled.org/core/api/tangled"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// inheritFromScm retrieve missing tangled settings from the tangled scm object.
func (t *Tangled) inheritFromScm() {
	if t.scm != nil {
		_, t.SourceBranch, t.TargetBranch = t.scm.GetBranches()
		t.Knot = t.scm.Spec.Knot
		t.Owner = t.scm.Spec.Owner
		t.Repository = t.scm.Spec.Repository
	}

	if t.spec.SourceBranch != "" {
		t.SourceBranch = t.spec.SourceBranch
	}
	if t.spec.TargetBranch != "" {
		t.TargetBranch = t.spec.TargetBranch
	}
	if t.spec.Knot != "" {
		t.Knot = t.spec.Knot
	}
	if t.spec.Owner != "" {
		t.Owner = t.spec.Owner
	}
	if t.spec.Repository != "" {
		t.Repository = t.spec.Repository
	}
}

// repoOverridesScm reports whether the PR spec points at a different
// repository than the attached SCM. When true, the scm-cached repoDid must
// not be used because it identifies the wrong repository.
func (t *Tangled) repoOverridesScm() bool {
	if t.scm == nil {
		return false
	}
	if t.spec.Owner != "" && t.spec.Owner != t.scm.Spec.Owner {
		return true
	}
	if t.spec.Repository != "" && t.spec.Repository != t.scm.Spec.Repository {
		return true
	}
	return false
}

// resolveTargetRepoDID returns the DID of the target repository as minted by
// the knot. The appview matches sh.tangled.repo.pull records against this DID.
func (t *Tangled) resolveTargetRepoDID(ctx context.Context) (string, error) {
	if t.spec.RepoDID != "" {
		return t.spec.RepoDID, nil
	}
	if !t.repoOverridesScm() && t.scm != nil {
		if did := t.scm.RepoDid(); did != "" {
			return did, nil
		}
	}
	if t.Owner == "" {
		return "", fmt.Errorf("cannot resolve target repository: no owner configured")
	}

	ownerDid := strings.TrimPrefix(t.Owner, "@")
	if !strings.HasPrefix(ownerDid, "did:") {
		resolved, err := t.client.ResolveHandle(ctx, ownerDid)
		if err != nil {
			return "", fmt.Errorf("resolve owner handle %q: %w", ownerDid, err)
		}
		ownerDid = resolved
	}

	records, err := t.client.ListRecords(ctx, ownerDid, tangled.RepoNSID)
	if err != nil {
		return "", fmt.Errorf("list %s records on %s: %w", tangled.RepoNSID, ownerDid, err)
	}

	want := strings.TrimSpace(t.Repository)
	for _, rec := range records {
		raw, err := json.Marshal(rec.Value)
		if err != nil {
			continue
		}
		var repo tangled.Repo
		if err := json.Unmarshal(raw, &repo); err != nil {
			continue
		}

		rkey := rec.Uri[strings.LastIndex(rec.Uri, "/")+1:]
		name := rkey
		if repo.Name != nil && *repo.Name != "" {
			name = *repo.Name
		}

		if name != want && rkey != want {
			continue
		}

		if repo.RepoDid != nil && *repo.RepoDid != "" {
			return *repo.RepoDid, nil
		}
		return "", fmt.Errorf("sh.tangled.repo record %s has no repoDid; the knot has not minted one yet", rec.Uri)
	}

	return "", fmt.Errorf("no sh.tangled.repo record on %s matched %q", ownerDid, want)
}

// generateFormatPatch runs `git format-patch` between target and source branches.
func (t *Tangled) generateFormatPatch() (string, error) {
	if t.scm == nil {
		return "", fmt.Errorf("format-patch generation requires a Tangled scm context")
	}
	dir := t.scm.GetDirectory()
	if dir == "" {
		return "", fmt.Errorf("scm working directory is empty")
	}
	revRange := fmt.Sprintf("%s..%s", t.TargetBranch, t.SourceBranch)
	cmd := exec.Command("git", "-C", dir, "format-patch", "--stdout", revRange)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git format-patch %s failed: %w: %s", revRange, err, stderr.String())
	}
	return stdout.String(), nil
}

// gzipPatch gzip-compresses a format-patch payload.
func gzipPatch(patch string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(patch)); err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// existingPull holds an open Tangled PR record that matches the current
// source/target combination. The caller uses it to append a new rounds[]
// entry instead of opening a duplicate PR.
type existingPull struct {
	Rkey   string
	Record tangled.RepoPull
	Link   string
}

// findExistingPull returns an open pull request matching the configured
// source/target repo+branch tuple. Tangled tracks open/closed/merged state
// in separate sh.tangled.repo.pull.status records: the latest such record
// (by rkey, which is a TID) wins. Pulls whose latest status is closed or
// merged are treated as absent so a replacement PR can be opened. Returns
// nil when no matching open pull exists.
func (t *Tangled) findExistingPull(ctx context.Context, targetRepoDid, sourceRepoDid string) (*existingPull, error) {
	ownerDid, err := t.client.ResolvedDID(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve owner DID: %w", err)
	}

	records, err := t.client.ListRecords(ctx, ownerDid, tangled.RepoPullNSID)
	if err != nil {
		return nil, fmt.Errorf("list pull records: %w", err)
	}

	statuses, err := t.latestPullStatuses(ctx, ownerDid)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		raw, err := json.Marshal(rec.Value)
		if err != nil {
			continue
		}
		var pr tangled.RepoPull
		if err := json.Unmarshal(raw, &pr); err != nil {
			logrus.Debugf("tangled: skipping unparseable pull record %s: %v", rec.Uri, err)
			continue
		}
		if pr.Target == nil {
			continue
		}
		if pr.Target.Repo != targetRepoDid || pr.Target.Branch != t.TargetBranch {
			continue
		}
		if pr.Source == nil || pr.Source.Branch != t.SourceBranch {
			continue
		}
		// Source.Repo may be nil for same-repo PRs; treat that as the target
		// repo for comparison so cross-repo and same-repo PRs with identical
		// branch names don't collide.
		recordSourceRepo := targetRepoDid
		if pr.Source.Repo != nil && *pr.Source.Repo != "" {
			recordSourceRepo = *pr.Source.Repo
		}
		if recordSourceRepo != sourceRepoDid {
			continue
		}

		switch statuses[rec.Uri] {
		case tangled.RepoPullStatusClosed, tangled.RepoPullStatusMerged:
			continue
		}

		link := fmt.Sprintf("%s/%s/%s/pulls", t.client.Appview(), t.Owner, t.Repository)
		logrus.Infof("%s Tangled pullrequest detected at:\n\t%s", result.SUCCESS, link)
		return &existingPull{
			Rkey:   rkeyFromURI(rec.Uri),
			Record: pr,
			Link:   link,
		}, nil
	}

	return nil, nil
}

// resolveSourceRepoDID returns the repoDid of the repository whose branch is
// being proposed. With an SCM attached this is the SCM's repoDid (the place
// where the working branch lives). When the SCM repoDid is unknown, the
// source falls back to the target repo so same-repo PRs still validate.
func (t *Tangled) resolveSourceRepoDID(targetRepoDid string) string {
	if t.scm != nil {
		if did := t.scm.RepoDid(); did != "" {
			return did
		}
	}
	return targetRepoDid
}

// latestPullStatuses returns the most recent status NSID for every pull
// at-uri found on the given owner's PDS. TIDs sort lexicographically by
// creation time, so the highest rkey wins.
func (t *Tangled) latestPullStatuses(ctx context.Context, ownerDid string) (map[string]string, error) {
	records, err := t.client.ListRecords(ctx, ownerDid, tangled.RepoPullStatusNSID)
	if err != nil {
		return nil, fmt.Errorf("list pull status records: %w", err)
	}

	latestRkey := map[string]string{}
	latestStatus := map[string]string{}
	for _, rec := range records {
		raw, err := json.Marshal(rec.Value)
		if err != nil {
			continue
		}
		var st tangled.RepoPullStatus
		if err := json.Unmarshal(raw, &st); err != nil {
			continue
		}
		if st.Pull == "" {
			continue
		}
		rkey := rkeyFromURI(rec.Uri)
		if rkey <= latestRkey[st.Pull] {
			continue
		}
		latestRkey[st.Pull] = rkey
		latestStatus[st.Pull] = st.Status
	}
	return latestStatus, nil
}

// rkeyFromURI extracts the rkey segment of an at-uri.
func rkeyFromURI(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
