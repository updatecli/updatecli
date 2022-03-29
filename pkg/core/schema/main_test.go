package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
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
    "$schema": "http://json-schema.org/draft/2020-12/schema",
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defer func() error {
		err := CleanCommentDirectory()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return nil
	}()

	err = s.GenerateSchema(&mockConfig{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if expectedJsonSchema != s.String() {
		t.Errorf("Expected Jsonschema:\n%s\nGot:%s",
			expectedJsonSchema,
			string(s.String()))
	}

}

func TestGenerateJsonSchema(t *testing.T) {
	expectedJsonSchema := `{
    "oneOf": [
        {
            "$schema": "http://json-schema.org/draft/2020-12/schema",
            "properties": {
                "sourceid": {
                    "type": "string"
                },
                "disablesourceinput": {
                    "type": "boolean"
                },
                "spec": {
                    "$schema": "http://json-schema.org/draft/2020-12/schema",
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	defer func() error {
		err := CleanCommentDirectory()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return nil
	}()

	anyOfSpec := map[string]interface{}{
		"jenkins": mockJenkinsSpec{},
	}

	schema := GenerateJsonSchema(mockConditionConfig{}, anyOfSpec)

	gotJsonSchema, err := json.MarshalIndent(schema, "", "    ")

	if err != nil {
		logrus.Errorf(err.Error())
	}

	if expectedJsonSchema != string(gotJsonSchema) {
		t.Errorf("Expected Jsonschema:\n%s\nGot:%s",
			expectedJsonSchema,
			string(gotJsonSchema))
	}
}

func TestGetPackageComments(t *testing.T) {
	for _, path := range []string{"../../.."} {
		comments, err := GetPackageComments(path)

		if err != nil {
			t.Errorf("Unexpected Error for path %q: %v", path, err)
		}

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
				// For simplifying error messag only show comments related to our test case
				if strings.HasPrefix(key, "github.com/updatecli/updatecli/pkg/core/config") {
					fmt.Printf("Debugging %q\n", key)
				}
			}
			t.Errorf("Unexpected result for path %q", path)
		}
	}
}
