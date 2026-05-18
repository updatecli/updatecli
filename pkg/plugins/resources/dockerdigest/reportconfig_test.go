package dockerdigest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportConfig_IncludesHideTag(t *testing.T) {
	source := &DockerDigest{spec: Spec{
		Image:   "registry.example.com/myapp",
		Tag:     "v1.0.0",
		HideTag: true,
	}}

	data, err := json.Marshal(source.ReportConfig())
	require.NoError(t, err)
	assert.Contains(t, string(data), `"HideTag":true`,
		"ReportConfig() must include HideTag so sources differing only in hidetag get distinct cache keys")
}
