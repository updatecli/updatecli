package gitea

import (
	"slices"
	"sync"

	"github.com/drone/go-scm/scm"
	"github.com/updatecli/updatecli/pkg/core/log"
)

var (
	// catalog is an in memory cache of published releases, keyed by
	// Gitea URL, owner and repository, to avoid repeated API calls.
	catalog map[string]*catalogEntry
	// catalogMu guards the catalog map, not the entries it holds.
	catalogMu sync.Mutex
)

// catalogEntry holds the cached releases of one repository. Its own lock
// lets fetches for different repositories run at the same time.
type catalogEntry struct {
	mu       sync.Mutex
	loaded   bool
	releases []*scm.Release
	// missing records tags absent after a refetch, so that asking for them
	// again does not refetch the whole history each time.
	missing map[string]bool
	// err records a failed first fetch, so that a repository that cannot be
	// listed is not requested again for the rest of the run.
	err error
}

// publishedReleases returns the repository's non-draft releases, newest first.
// The list is cached per repository so that several targets sharing a source
// do not fetch the whole release history again. If the cached list does not
// contain the release "to", it is fetched again once, since that release may
// have been published after the cache was filled.
func publishedReleases(id, to string, fetch func() ([]*scm.Release, error)) ([]*scm.Release, error) {
	entry := catalogEntryFor(id)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.err != nil {
		return nil, entry.err
	}

	hasTo := func() bool {
		return to == "" || slices.ContainsFunc(entry.releases, func(r *scm.Release) bool { return r.Tag == to })
	}

	if entry.loaded && (hasTo() || entry.missing[to]) {
		return entry.releases, nil
	}

	if entry.missing == nil {
		entry.missing = map[string]bool{}
	}

	if entry.loaded {
		log.Debugf("Gitea release %q not found in cached releases, fetching them again", to)
	}

	releases, err := fetch()
	if err != nil {
		if entry.loaded {
			// Keep the earlier list rather than failing a changelog that may not need "to".
			log.Debugf("failed to refresh Gitea releases, using cached ones: %s", err)
			entry.missing[to] = true
			return entry.releases, nil
		}
		entry.err = err
		return nil, err
	}
	entry.releases = slices.DeleteFunc(releases, func(r *scm.Release) bool { return r.Draft })
	entry.loaded = true
	if !hasTo() {
		entry.missing[to] = true
	}

	return entry.releases, nil
}

// catalogEntryFor returns the catalog entry for id, creating it if needed.
func catalogEntryFor(id string) *catalogEntry {
	catalogMu.Lock()
	defer catalogMu.Unlock()

	if catalog == nil {
		catalog = make(map[string]*catalogEntry)
	}
	entry, ok := catalog[id]
	if !ok {
		entry = &catalogEntry{}
		catalog[id] = entry
	}
	return entry
}
