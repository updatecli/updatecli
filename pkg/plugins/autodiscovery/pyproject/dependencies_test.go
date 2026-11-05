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

func TestDependencyGroupMemberships(t *testing.T) {
	groups := map[string]interface{}{
		"Test_Group": []interface{}{"pytest>=8.0"},
		"lint":       []interface{}{"ruff>=0.4"},
		// Included name written differently than declared, compared normalized
		"dev": []interface{}{map[string]interface{}{"include-group": "test-group"}, "mypy"},
		// Array of tables syntax, including dev which includes Test_Group
		"all": []map[string]interface{}{{"include-group": "dev"}, {"include-group": "lint"}},
		// Cycle, forbidden by PEP 735, must not loop
		"a": []interface{}{map[string]interface{}{"include-group": "b"}},
		"b": []interface{}{map[string]interface{}{"include-group": "a"}},
	}

	memberships := dependencyGroupMemberships(groups)

	assert.Equal(t, []string{"Test_Group", "all", "dev"}, memberships["Test_Group"])
	assert.Equal(t, []string{"all", "lint"}, memberships["lint"])
	assert.Equal(t, []string{"all", "dev"}, memberships["dev"])
	assert.Equal(t, []string{"all"}, memberships["all"])
	assert.Equal(t, []string{"a", "b"}, memberships["a"])
	assert.Equal(t, []string{"a", "b"}, memberships["b"])
}
