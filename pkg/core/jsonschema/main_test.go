package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

// mockConditionConfig defines conditions input parameters
// the motivation is to avoid circular import in testing
type mockConditionConfig struct {
	//resource.ResourceConfig `yaml:",inline"`
	// SourceID defines which source is used to retrieve the default value
	SourceID string `yaml:"sourceID"`
	// DisableSourceInput allows to not retrieve default source value.
	DisableSourceInput bool
	Spec               interface{}
	Kind               string `jsonschema:"required"`
}

type mockJenkinsSpec struct {
	// Release defines a mock release
	Release string
}

// mockConfig contains cli configuration
// the motivation is to avoid circular import in testing
type mockConfig struct {
	Name string
	// PipelineID allows to identify a full pipeline run, this value is propagated into each target if not defined at that level
	PipelineID string
	// Title is used for the full pipeline
	Title string
	// Conditions defines the list of condition configuration
	Conditions map[string]mockConditionConfig
}

func TestGenerateSchema(t *testing.T) {
	expectedJsonSchema := `{
    "$schema": "http://json-schema.org/draft-04/schema",
    "$id": "https://www.updatecli.io/schema/mock-config",
    "properties": {
        "name": {
            "type": "string"
        },
        "pipelineid": {
            "type": "string"
        },
        "title": {
            "type": "string"
        },
        "conditions": {
            "patternProperties": {
                ".*": {
                    "properties": {
                        "sourceid": {
                            "type": "string"
                        },
                        "disablesourceinput": {
                            "type": "boolean"
                        },
                        "spec": true,
                        "kind": {
                            "type": "string"
                        }
                    },
                    "additionalProperties": false,
                    "type": "object",
                    "required": [
                        "kind"
                    ]
                }
            },
            "type": "object"
        }
    },
    "additionalProperties": false,
    "type": "object"
}`
	s := New("", "")

	err := CloneCommentDirectory()
	require.NoError(t, err)

	defer func() {
		err := CleanCommentDirectory()
		require.NoError(t, err)
	}()

	err = s.GenerateSchema(&mockConfig{})
	require.NoError(t, err)

	assert.Equal(t, expectedJsonSchema, s.String())
}

func TestGenerateJsonSchema(t *testing.T) {
	expectedJsonSchema := `{
    "oneOf": [
        {
            "$schema": "http://json-schema.org/draft-04/schema",
            "properties": {
                "sourceid": {
                    "type": "string"
                },
                "disablesourceinput": {
                    "type": "boolean"
                },
                "spec": {
                    "$schema": "http://json-schema.org/draft-04/schema",
                    "properties": {
                        "release": {
                            "type": "string"
                        }
                    },
                    "additionalProperties": false,
                    "type": "object"
                },
                "kind": {
                    "enum": [
                        "jenkins"
                    ]
                }
            },
            "additionalProperties": false,
            "type": "object",
            "required": [
                "kind"
            ]
        }
    ]
}`
	err := CloneCommentDirectory()
	require.NoError(t, err)

	defer func() {
		err := CleanCommentDirectory()
		require.NoError(t, err)
	}()

	anyOfSpec := map[string]interface{}{
		"jenkins": mockJenkinsSpec{},
	}

	schema := AppendOneOfToJsonSchema(mockConditionConfig{}, anyOfSpec)

	gotJsonSchema, err := json.MarshalIndent(schema, "", "    ")
	require.NoError(t, err)

	assert.Equal(t, expectedJsonSchema, string(gotJsonSchema))
}

func TestGetPackageComments(t *testing.T) {
	for _, path := range []string{"../../.."} {
		comments, err := GetPackageComments(path)
		require.NoError(t, err)

		expectedResult := false
		for key := range comments {
			// Testing arbitrary comment that should exist
			if strings.HasPrefix(key, "github.com/updatecli/updatecli/pkg/core/config.Config") {
				expectedResult = true
				break
			}
		}

		if !expectedResult {
			for key := range comments {
				// To simplify error message, it only show comments related to our test case
				if strings.HasPrefix(key, "github.com/updatecli/updatecli/pkg/core/config") {
					fmt.Printf("Debugging %q\n", key)
				}
			}
			t.Errorf("Unexpected result for path %q", path)
		}
	}
}

func TestGenerateSpecToMapJsonSchema(t *testing.T) {

	type emptyMap map[string]interface{}

	dataset := []struct {
		expectedJsonSchema string
		baseSchema         interface{}
		input              map[string]interface{}
	}{
		{
			input: map[string]interface{}{
				"jenkins": mockJenkinsSpec{},
			},
			baseSchema: emptyMap{},
			expectedJsonSchema: `{
    "$schema": "http://json-schema.org/draft-04/schema",
    "properties": {
        "jenkins": {
            "$schema": "http://json-schema.org/draft-04/schema",
            "properties": {
                "release": {
                    "type": "string"
                }
            },
            "additionalProperties": false,
            "type": "object"
        }
    },
    "type": "object"
}`},
		{
			input: map[string]interface{}{
				"jenkins": mockJenkinsSpec{},
			},
			baseSchema: mockConditionConfig{},
			expectedJsonSchema: `{
    "$schema": "http://json-schema.org/draft-04/schema",
    "properties": {
        "sourceid": {
            "type": "string"
        },
        "disablesourceinput": {
            "type": "boolean"
        },
        "spec": true,
        "kind": {
            "type": "string"
        },
        "jenkins": {
            "$schema": "http://json-schema.org/draft-04/schema",
            "properties": {
                "release": {
                    "type": "string"
                }
            },
            "additionalProperties": false,
            "type": "object"
        }
    },
    "additionalProperties": false,
    "type": "object",
    "required": [
        "kind"
    ]
}`},
	}

	err := CloneCommentDirectory()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defer func() {
		err := CleanCommentDirectory()
		if err != nil {
			t.Errorf("unexpected error while cleaning comment directory: %v", err)
		}
	}()

	for _, data := range dataset {
		schema := AppendMapToJsonSchema(data.baseSchema, data.input)

		gotJsonSchema, err := json.MarshalIndent(schema, "", "    ")
		require.NoError(t, err)

		assert.Equal(t, data.expectedJsonSchema, string(gotJsonSchema))
	}
}
