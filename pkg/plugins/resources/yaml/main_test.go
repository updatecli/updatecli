package yaml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Validate(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		mockFileExist bool
		isErrorWanted bool
	}{
		{
			name: "Normal case with 'File'",
			spec: Spec{
				File: "/tmp/test.yaml",
				Key:  "foo.bar",
			},
			isErrorWanted: false,
		},
		{
			name: "Normal case with more than one 'Files'",
			spec: Spec{
				Files: []string{
					"/tmp/test.yaml",
					"/tmp/bar.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: false,
		},
		{
			name: "Validation error when both 'File' and 'Files' are empty",
			spec: Spec{
				File:  "",
				Files: []string{},
				Key:   "foo.bar",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when 'Key' is empty",
			spec: Spec{
				File: "/tmp/toto.yaml",
				Key:  "",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when both 'File' and 'Files' are specified",
			spec: Spec{
				File: "test.yaml",
				Files: []string{
					"bar.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: true,
		},
		{
			name: "Validation error when 'Files' contains duplicates",
			spec: Spec{
				Files: []string{
					"test.yaml",
					"test.yaml",
				},
				Key: "foo.bar",
			},
			isErrorWanted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := Yaml{
				spec: tt.spec,
			}
			gotErr := yaml.spec.Validate()
			if tt.isErrorWanted {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
