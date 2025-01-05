package config

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
)

// TestTemplateData contains all the data required to run a test
type TestTemplateData struct {
	// ID is used to identify the test
	ID string
	// ManifestTemplate is the template to render
	ManifestTemplate string
	// Values1 contains key/value extracted from the first values file
	Values1 string
	// Values2 would always be call after Values1 so values from Values2 will override values from Values1
	Values2 string
	// ExpectedManifest is the expected manifest after rendering the template
	ExpectedManifest string
	// ExpectError is used to check if the test should return an error
	ExpectError bool
}

var (
	// testTemplates contains all the test cases
	testTemplates = []TestTemplateData{
		{
			ID: "render value",
			ManifestTemplate: `
      hello {{ .value }}
      `,
			Values1: `
      value: my friend
      `,
			Values2: `
      value: world
      `,
			ExpectedManifest: `
      hello world
      `,
		},

		{
			ID: "use sprig template function",
			ManifestTemplate: `
      hello {{ .value | default "world" .value }}
      `,
			ExpectedManifest: `
      hello world
      `,
		},

		{
			ID: "invalid template",
			ManifestTemplate: `
      hello {{ .value }
      `,
			ExpectError: true,
		},

		{
			ID: "built-in function",
			ManifestTemplate: `
      hello {{ requiredEnv "FOO" }}
      `,
			ExpectedManifest: `
      hello BAR
      `,
		},
	}
)

func TestTemplates(t *testing.T) {
	t.Setenv("FOO", "BAR")
	for _, testTemplate := range testTemplates {
		t.Run(fmt.Sprintf("test template %s", testTemplate.ID), func(t *testing.T) {

			template := Template{
				ValuesFiles: []string{"values1.yml", "values2.yml"},
				fs: fstest.MapFS{
					"values1.yml": {Data: []byte(testTemplate.Values1)},
					"values2.yml": {Data: []byte(testTemplate.Values2)},
				},
			}
			expected := dedent.Dedent(testTemplate.ExpectedManifest)
			rendered, err := template.New([]byte(dedent.Dedent(testTemplate.ManifestTemplate)))
			if testTemplate.ExpectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expected, string(rendered))
		})
	}

	t.Run("missing values file", func(t *testing.T) {
		template := Template{
			ValuesFiles: []string{"values.yml"},
			fs:          fstest.MapFS{},
		}
		_, err := template.New([]byte(""))
		require.Equal(t, err.Error(), fmt.Sprintf("open values.yml: %s", fs.ErrNotExist.Error()))
	})
}

func TestMergeValueFile(t *testing.T) {

	testdata := []struct {
		name          string
		data1         map[string]interface{}
		data2         map[string]interface{}
		expectedValue map[string]interface{}
	}{
		{
			name: "merge values",
			data1: map[string]interface{}{
				"key1": "value1",
			},
			data2: map[string]interface{}{
				"key2": "value2",
			},
			expectedValue: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "no merge values",
			data1: map[string]interface{}{
				"key1": "value1",
			},
			expectedValue: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "overridden values",
			data1: map[string]interface{}{
				"key1": "value1",
				"key2": "value3",
			},
			data2: map[string]interface{}{
				"key2": "value2",
			},
			expectedValue: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:          "no values",
			data1:         map[string]interface{}{},
			data2:         map[string]interface{}{},
			expectedValue: map[string]interface{}{},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedValue, mergeValueFile(tt.data1, tt.data2))
		})
	}
}
