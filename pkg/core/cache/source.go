package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
)

// SourceEntry stores the cached result of a source execution.
// Transformers are not included — they are re-applied on cache hit so that
// sources sharing the same underlying resource config but different transformer
// chains each get the correctly transformed output.
type SourceEntry struct {
	Information string // Raw value returned by the plugin, before transformers
	Description string
	Result      string // SUCCESS / FAILURE / etc.
}

// SourceCache is a thread-safe in-memory cache for source execution results,
// keyed by SHA256 of the sanitized resource configuration.
// The cache lives for the duration of one updatecli execution and is shared
// across all pipelines, allowing identical sources to be executed only once.
//
// NOTE: today pipelines run sequentially and DAG nodes are serialized by a
// mutex, so concurrent cache misses for the same key cannot happen. If pipeline
// execution is ever parallelized, consider adding a singleflight.Group to
// coalesce concurrent lookups for the same key.
type SourceCache struct {
	mu      sync.RWMutex
	entries map[string]SourceEntry
}

// NewSourceCache creates a new empty source cache.
func NewSourceCache() *SourceCache {
	return &SourceCache{
		entries: make(map[string]SourceEntry),
	}
}

// cacheKeyInput contains only the fields that determine what a source plugin
// returns. Name, Transformers, DependsOn, and SCMID are excluded so that
// sources with identical plugin config share the same cache entry regardless
// of how they are named or wired in different pipelines.
type cacheKeyInput struct {
	Kind string `json:"kind"`
	Spec any    `json:"spec"`
}

// Key computes a cache key from a ResourceConfig by hashing only the Kind and
// sanitized Spec (via ReportConfig). This instantiates the resource plugin
// internally; on a cache miss the caller will instantiate it again for
// execution. Returns empty string when hashing fails; callers treat that as a
// cache miss.
func Key(rc resource.ResourceConfig) string {
	r, err := resource.New(rc)
	if err != nil {
		logrus.Debugf("source cache: failed to instantiate resource for key: %v", err)
		return ""
	}

	data, err := json.Marshal(cacheKeyInput{
		Kind: rc.Kind,
		Spec: r.ReportConfig(),
	})
	if err != nil {
		logrus.Debugf("source cache: failed to marshal config for key: %v", err)
		return ""
	}

	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// Get retrieves a cached source entry. Returns the entry and true if found.
func (c *SourceCache) Get(key string) (SourceEntry, bool) {
	if key == "" {
		return SourceEntry{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	return entry, ok
}

// Set stores a source entry in the cache.
func (c *SourceCache) Set(key string, entry SourceEntry) {
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry
}

// Len returns the number of entries currently held in the cache.
func (c *SourceCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
