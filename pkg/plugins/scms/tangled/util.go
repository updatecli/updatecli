package tangled

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"tangled.org/core/api/tangled"
)

func (t *Tangled) setDirectory() {
	if _, err := os.Stat(t.Spec.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(t.Spec.Directory, 0o755); err != nil {
			logrus.Errorf("err - %s", err)
		}
	}
}

// resolveRepoRecord queries the owner's PDS for the sh.tangled.repo record
// that matches t.Spec.Repository and caches the knot hostname and repoDid.
func (t *Tangled) resolveRepoRecord(ctx context.Context) error {
	ownerDid, err := t.resolveOwnerDID(ctx)
	if err != nil {
		return err
	}

	records, err := t.client.ListRecords(ctx, ownerDid, tangled.RepoNSID)
	if err != nil {
		return fmt.Errorf("list repo records: %w", err)
	}

	want := strings.TrimSpace(t.Spec.Repository)
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

		if t.Spec.Knot == "" {
			t.Spec.Knot = repo.Knot
		}
		if repo.RepoDid != nil {
			t.repoDid = *repo.RepoDid
		}
		logrus.Debugf("tangled: resolved knot=%q repoDid=%q for %s/%s", t.Spec.Knot, t.repoDid, t.Spec.Owner, want)
		return nil
	}

	return fmt.Errorf("no sh.tangled.repo record on %s matched %q", ownerDid, want)
}

// resolveOwnerDID returns the DID for the configured owner, resolving handles via the identity directory.
func (t *Tangled) resolveOwnerDID(ctx context.Context) (string, error) {
	owner := strings.TrimPrefix(t.Spec.Owner, "@")
	if owner == "" {
		return "", fmt.Errorf("owner is empty")
	}
	if strings.HasPrefix(owner, "did:") {
		return owner, nil
	}
	return t.client.ResolveHandle(ctx, owner)
}

// RepoDid returns the knot-minted DID for this repository when known.
func (t *Tangled) RepoDid() string {
	return t.repoDid
}
