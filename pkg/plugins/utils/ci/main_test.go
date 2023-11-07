package ci

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	tests := []struct {
		name    string
		env     map[string]string
		expect  CIEngine
		wantErr error
	}{
		// Note we do not test the "unknown" case as the current CI environment might have its own variables already set
		// It could also have an impact on the test below (depends on the code order)
		{
			name: "Jenkins",
			env: map[string]string{
				"JENKINS_URL": "http://example.com",
			},
			expect: Jenkins{},
		},
		{
			name: "GitLab CI",
			env: map[string]string{
				"GITLAB_CI": "http://example.com",
			},
			expect: GitLabCi{},
		},
		{
			name: "GitHub Actions",
			env: map[string]string{
				"GITHUB_ACTION": "http://example.com",
			},
			expect: GitHubActions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, gotErr := New()

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expect, got)
		})
	}
}
