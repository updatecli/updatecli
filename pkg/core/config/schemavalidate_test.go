package config

import (
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schemaValidationTestCase struct {
	name string
	// manifest holds the manifest content, indented for readability and dedented before
	// being validated.
	manifest string
	options  *SchemaValidationOptions
	// expectedErrors and expectedWarnings are matched as substrings against the problems
	// of the matching severity, which must all be accounted for.
	expectedErrors   []string
	expectedWarnings []string
	expectedParseErr bool
}

func TestValidateSchema(t *testing.T) {

	strict := DefaultSchemaValidationOptions()
	strict.Strict = true

	testCases := []schemaValidationTestCase{
		{
			name: "a valid manifest",
			manifest: `
				name: a valid manifest
				sources:
				  default:
				    kind: shell
				    spec:
				      command: echo 1
				targets:
				  default:
				    kind: file
				    sourceid: default
				    spec:
				      file: /tmp/example
				`,
		},
		{
			name: "a manifest without a name, which defaults to the filename",
			manifest: `
				sources:
				  default:
				    kind: shell
				    spec:
				      command: echo 1
				`,
		},
		{
			name: "an unknown top level key",
			manifest: `
				name: a manifest
				sourcs:
				  default:
				    kind: shell
				`,
			expectedErrors: []string{`unknown key "sourcs", did you mean "sources"?`},
		},
		{
			name: "an unknown resource level key",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: shell
				    scmid_typo: default
				    spec:
				      command: echo 1
				`,
			expectedErrors: []string{`sources.default: unknown key "scmid_typo"`},
		},
		{
			// The headline case: a typo inside a spec must name the offending key rather
			// than report that the resource matched none of the supported kinds.
			name: "an unknown key inside a spec",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: shell
				    spec:
				      commnd: echo 1
				`,
			expectedErrors: []string{`sources.default.spec: unknown key "commnd", did you mean "command"?`},
			// Typing a key wrong also leaves the key it was meant to be missing.
			expectedWarnings: []string{`sources.default.spec: missing required key(s) "command"`},
		},
		{
			name: "an unknown kind",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: dockerimag
				    spec:
				      image: updatecli/updatecli
				`,
			expectedErrors: []string{`sources.default.kind: unsupported kind "dockerimag", did you mean "dockerimage"?`},
		},
		{
			name: "a missing kind",
			manifest: `
				name: a manifest
				sources:
				  default:
				    spec:
				      command: echo 1
				`,
			expectedErrors: []string{`sources.default: missing required key "kind"`},
		},
		{
			// Updatecli lowercases a kind before dispatching on it.
			name: "a kind that is not lowercase",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: Shell
				    spec:
				      command: echo 1
				`,
			expectedWarnings: []string{`kind value "Shell" must be lowercase`},
		},
		{
			// A spec is decoded with mapstructure, which ignores the case of its keys.
			name: "a spec key that is not lowercase",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: dockerimage
				    spec:
				      Image: updatecli/updatecli
				      VersionFilter:
				        Kind: semver
				`,
		},
		{
			name: "a deprecated key",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: shell
				    scmID: default
				    spec:
				      command: echo 1
				scms:
				  default:
				    kind: git
				    spec:
				      url: https://example.com/repository.git
				`,
			expectedWarnings: []string{`"scmID" is deprecated in favor of "scmid"`},
		},
		{
			name: "a deprecated key reported as an error when strict",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: shell
				    scmID: default
				    spec:
				      command: echo 1
				scms:
				  default:
				    kind: git
				    spec:
				      url: https://example.com/repository.git
				`,
			options:        &strict,
			expectedErrors: []string{`"scmID" is deprecated in favor of "scmid"`},
		},
		{
			// A template is only resolved once the pipeline runs, so the value reaching
			// the validator is always a string whatever the field type.
			name: "a templated value on a non string field",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: gittag
				    spec:
				      lsremote: '{{ source "enabled" }}'
				`,
		},
		{
			name: "a templated value reported when strict",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: gittag
				    spec:
				      lsremote: '{{ source "enabled" }}'
				`,
			options:        &strict,
			expectedErrors: []string{`sources.default.spec.lsremote: expected boolean, got string`},
		},
		{
			name: "a wrong type",
			manifest: `
				name: a manifest
				sources:
				  default:
				    kind: shell
				    spec:
				      command: 1
				`,
			expectedErrors: []string{`sources.default.spec.command: expected string, got number`},
		},
		{
			// An scm may be disabled, in which case it carries neither kind nor spec.
			name: "a disabled scm without a kind",
			manifest: `
				name: a manifest
				scms:
				  default:
				    disabled: true
				`,
		},
		{
			name: "a resource that is not a mapping",
			manifest: `
				name: a manifest
				sources:
				  default: a string
				`,
			expectedErrors: []string{`sources.default: expected a mapping`},
		},
		{
			name: "a manifest that is not valid yaml",
			manifest: `
				name: a manifest
				sources:
				 default:
				   kind: shell
				  spec: {}
				`,
			expectedParseErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			options := DefaultSchemaValidationOptions()
			if testCase.options != nil {
				options = *testCase.options
			}

			report, err := ValidateSchema(
				"manifest.yaml", []byte(dedent.Dedent(testCase.manifest)), options)

			if testCase.expectedParseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assertProblems(t, testCase.expectedErrors, report.Errors(), "error")
			assertProblems(t, testCase.expectedWarnings, report.Warnings(), "warning")
		})
	}
}

// assertProblems checks that each expected substring matches one reported problem, and
// that no problem is left unaccounted for.
func assertProblems(t *testing.T, expected []string, got []SchemaProblem, severity string) {
	t.Helper()

	messages := make([]string, len(got))
	for i, problem := range got {
		messages[i] = problem.String()
	}

	require.Lenf(t, got, len(expected), "unexpected %s(s): %v", severity, messages)

	for _, want := range expected {
		matched := false
		for _, message := range messages {
			if strings.Contains(message, want) {
				matched = true
				break
			}
		}
		assert.Truef(t, matched, "no %s matching %q in %v", severity, want, messages)
	}
}

// TestValidateSchemaMultiDocument ensures each document of a manifest is validated and
// that a problem names the document it was found in.
func TestValidateSchemaMultiDocument(t *testing.T) {

	manifest := dedent.Dedent(`
		name: the first document
		sources:
		  default:
		    kind: shell
		    spec:
		      command: echo 1
		---
		name: the second document
		sources:
		  default:
		    kind: shell
		    spec:
		      commnd: echo 2
		`)

	report, err := ValidateSchema("manifest.yaml", []byte(manifest), DefaultSchemaValidationOptions())
	require.NoError(t, err)

	problems := report.Errors()
	require.Len(t, problems, 1)

	assert.Equal(t, 1, problems[0].Document)
	assert.Contains(t, problems[0].String(), "manifest.yaml[1]")
}
