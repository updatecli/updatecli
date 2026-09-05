package gitcommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    map[string]interface{}
		wantAge age.Spec
		wantErr []string
	}{
		{
			name: "path and branch",
			spec: map[string]interface{}{
				"path":   "/tmp/repository",
				"branch": "main",
			},
		},
		{
			name: "age filter",
			spec: map[string]interface{}{
				"path":   "/tmp/repository",
				"branch": "main",
				"age":    map[string]interface{}{"minimum": "7d", "maximum": "30d"},
			},
			wantAge: age.Spec{Minimum: "7d", Maximum: "30d"},
		},
		{
			name: "invalid age filter",
			spec: map[string]interface{}{
				"path":   "/tmp/repository",
				"branch": "main",
				"age":    map[string]interface{}{"minimum": "7 days"},
			},
			wantErr: []string{"validation error", `invalid MinimumReleaseAge "7 days"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := New(tt.spec)

			if len(tt.wantErr) > 0 {
				require.Error(t, err)
				for _, want := range tt.wantErr {
					assert.Contains(t, err.Error(), want)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "/tmp/repository", resource.spec.Path)
			assert.Equal(t, "main", resource.spec.Branch)
			assert.Equal(t, tt.wantAge, resource.spec.Age)
		})
	}
}

func TestReportConfig(t *testing.T) {
	depth := 1
	resource := &GitCommit{spec: Spec{
		Path:     "/tmp/repository",
		Branch:   "main",
		Age:      age.Spec{Minimum: "7d", Maximum: "30d"},
		Hash:     "abc123",
		Depth:    &depth,
		URL:      "https://user:secret@example.com/owner/repository.git",
		Username: "user",
		Password: "secret",
	}}

	assert.Equal(t, Spec{
		Path:   "/tmp/repository",
		Branch: "main",
		Age:    age.Spec{Minimum: "7d", Maximum: "30d"},
		Hash:   "abc123",
		Depth:  &depth,
		URL:    "https://****:****@example.com/owner/repository.git",
	}, resource.ReportConfig())
}
