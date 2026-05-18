package dockerdigest

// TestReportConfigCacheKey proves that ReportConfig() must include HideTag in
// its output so that two sources sharing the same image+tag but differing only
// in hidetag receive distinct cache keys.
//
// RED (bug):  HideTag absent  → identical JSON → same key → cache collision
// GREEN (fix): HideTag present → different JSON → different keys → no collision

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cacheKeyFromSpec(spec interface{}) string {
	payload, _ := json.Marshal(struct {
		Kind string      `json:"Kind"`
		Spec interface{} `json:"Spec"`
	}{Kind: "dockerdigest", Spec: spec})
	h := sha256.Sum256(payload)
	return fmt.Sprintf("%x", h)
}

func TestReportConfig_HideTagIncluded(t *testing.T) {
	const image = "registry.example.com/myapp"
	const tag = "v1.0.0"

	versionSource := &DockerDigest{spec: Spec{
		Image:   image,
		Tag:     tag,
		HideTag: false,
	}}
	appVersionSource := &DockerDigest{spec: Spec{
		Image:   image,
		Tag:     tag,
		HideTag: true,
	}}

	rcVersion := versionSource.ReportConfig()
	rcAppVersion := appVersionSource.ReportConfig()

	keyVersion := cacheKeyFromSpec(rcVersion)
	keyAppVersion := cacheKeyFromSpec(rcAppVersion)

	// Marshal both to inspect what fields are actually present.
	jsonVersion, err := json.Marshal(rcVersion)
	require.NoError(t, err)
	jsonAppVersion, err := json.Marshal(rcAppVersion)
	require.NoError(t, err)

	t.Logf("latestVersion    ReportConfig JSON: %s", jsonVersion)
	t.Logf("latestAppVersion ReportConfig JSON: %s", jsonAppVersion)
	t.Logf("latestVersion    cache key: %s", keyVersion)
	t.Logf("latestAppVersion cache key: %s", keyAppVersion)

	// GREEN: keys must differ so each source executes independently.
	assert.NotEqual(t, keyVersion, keyAppVersion,
		"cache key collision: HideTag must be included in ReportConfig() output "+
			"so latestVersion and latestAppVersion receive distinct cache keys")

	// Verify HideTag is actually present in the appVersion JSON.
	assert.Contains(t, string(jsonAppVersion), "HideTag",
		"ReportConfig() must include HideTag field")
}
