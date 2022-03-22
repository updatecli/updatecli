package schema

import (
	"encoding/json"
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
	s := New("", "")
	schema := s.GenerateSchema(&mockConfig{})

	t.Errorf("Error: %v", schema)
}

func TestGetPackageComments(t *testing.T) {
	comments, err := GetPackageComments("")

	if err != nil {
		t.Errorf("Unexpected Error: %v", err)

	}

	expectedResult := false

	for key := range comments {
		if strings.HasPrefix(key, "github.com/updatecli/updatecli/pkg/core") {
			expectedResult = true
			break
		}
	}

	if !expectedResult {
		t.Errorf("Unexpected result: %v", comments)
	}

}

func TestGenerateJsonSchema(t *testing.T) {

	anyOfSpec := map[string]interface{}{
		"jenkins": mockJenkinsSpec{},
	}

	schema := GenerateJsonSchema(mockConditionConfig{}, anyOfSpec)

	u, err := json.MarshalIndent(schema, "", "    ")

	if err != nil {
		logrus.Errorf(err.Error())
	}

	t.Error(string(u))
}
