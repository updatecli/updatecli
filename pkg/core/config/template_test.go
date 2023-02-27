package config

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/require"
)

type TestTemplateData struct {
	ID               string
	ManifestTemplate string
	Values           string
	ExpectedManifest string
	ExpectError      bool
}

var (
	testTemplates = []TestTemplateData{
		{
			ID: "render value",
			ManifestTemplate: `
      hello {{ .value }}
      `,
			Values: `
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
				ValuesFiles: []string{"values.yml"},
				fs: fstest.MapFS{
					"values.yml": {Data: []byte(testTemplate.Values)},
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
