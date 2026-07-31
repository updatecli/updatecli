package compose

import (
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"go.yaml.in/yaml/v3"
)

func TestComposeSchemaValidation(t *testing.T) {

	schema, err := getComposeSchema()
	require.NoError(t, err)

	testCases := []struct {
		name string
		// compose holds the file content, indented for readability and dedented before
		// being validated.
		compose          string
		expectedProblems []string
	}{
		{
			name: "a valid compose file",
			compose: `
				name: a compose file
				policies:
				  - name: local
				    config:
				      - updatecli.d/example.yaml
				`,
		},
		{
			name: "a misspelled top level key",
			compose: `
				name: a compose file
				polices:
				  - name: local
				`,
			expectedProblems: []string{`unknown key "polices"`},
		},
		{
			name: "a misspelled policy key",
			compose: `
				name: a compose file
				policies:
				  - name: local
				    confg:
				      - updatecli.d/example.yaml
				`,
			expectedProblems: []string{`policies.0: unknown key "confg"`},
		},
		{
			name: "a wrong type",
			compose: `
				name: a compose file
				policies: a string
				`,
			expectedProblems: []string{`policies: expected array, got string`},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			var document interface{}
			require.NoError(t, yaml.Unmarshal([]byte(dedent.Dedent(testCase.compose)), &document))

			problems := jsonschema.Validate(schema, document)

			messages := make([]string, len(problems))
			for i, problem := range problems {
				messages[i] = problem.String()
			}

			require.Lenf(t, problems, len(testCase.expectedProblems),
				"unexpected problem(s): %v", messages)

			for _, want := range testCase.expectedProblems {
				matched := false
				for _, message := range messages {
					if strings.Contains(message, want) {
						matched = true
						break
					}
				}
				assert.Truef(t, matched, "no problem matching %q in %v", want, messages)
			}
		})
	}
}
