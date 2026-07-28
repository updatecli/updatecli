package jsonschema

import (
	"testing"

	jschema "github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSchemaID string = "https://www.updatecli.io/schema/test/condition"

func TestCompile(t *testing.T) {

	schemas := BuildKindSchemas(mockConditionConfig{}, map[string]interface{}{
		"jenkins": mockJenkinsSpec{},
	})
	require.Contains(t, schemas, "jenkins")

	compiled, err := Compile(testSchemaID, schemas["jenkins"])
	require.NoError(t, err)

	testCases := []struct {
		name        string
		instance    map[string]interface{}
		expectError bool
	}{
		{
			name: "a valid condition",
			instance: map[string]interface{}{
				"kind": "jenkins",
				"spec": map[string]interface{}{"release": "v1.0.0"},
			},
		},
		{
			name: "an unknown key in the spec",
			instance: map[string]interface{}{
				"kind": "jenkins",
				"spec": map[string]interface{}{"relase": "v1.0.0"},
			},
			expectError: true,
		},
		{
			name: "a wrong type in the spec",
			instance: map[string]interface{}{
				"kind": "jenkins",
				"spec": map[string]interface{}{"release": 1},
			},
			expectError: true,
		},
		{
			name: "a kind that does not match the pinned one",
			instance: map[string]interface{}{
				"kind": "dockerimage",
			},
			expectError: true,
		},
		{
			name:        "a missing required kind",
			instance:    map[string]interface{}{},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := compiled.Validate(testCase.instance)

			if testCase.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// TestCompileIgnoresSchemaVersion ensures validation does not depend on the invopop
// package global that New() mutates to draft-04, which would otherwise make the
// declared draft depend on whether a schema was exported earlier in the same process.
func TestCompileIgnoresSchemaVersion(t *testing.T) {

	kinds := map[string]interface{}{"jenkins": mockJenkinsSpec{}}

	previousVersion := jschema.Version
	defer func() { jschema.Version = previousVersion }()

	instance := map[string]interface{}{
		"kind": "jenkins",
		"spec": map[string]interface{}{"relase": "v1.0.0"},
	}

	for _, version := range []string{schemaVersionDraft04, "https://json-schema.org/draft/2020-12/schema"} {
		jschema.Version = version

		compiled, err := Compile(testSchemaID, BuildKindSchemas(mockConditionConfig{}, kinds)["jenkins"])
		require.NoErrorf(t, err, "compiling a schema declaring %q", version)

		assert.Errorf(t, compiled.Validate(instance),
			"an unknown spec key must be reported whatever the declared draft %q", version)
	}
}
