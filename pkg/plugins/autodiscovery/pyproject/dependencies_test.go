package pyproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A malformed dependency group must not hide the rest of the pyproject.toml,
// and a group written as an array of tables ("all") is valid and not reported as malformed.
// A group named "it's", which PEP 735 doesn't allow, is skipped.
func TestDiscoverManifestsMalformedDependencyGroup(t *testing.T) {
	p, err := New(Spec{}, "testdata/malformed_dependency_groups", "", "")
	require.NoError(t, err)

	manifests, err := p.DiscoverManifests()
	require.NoError(t, err)
	require.Len(t, manifests, 2)

	assert.Contains(t, string(manifests[0]), `bump "requests" for "malformed-groups-project" project`)
	assert.Contains(t, string(manifests[1]), `bump "pytest" [group:dev] for "malformed-groups-project" project`)
}
