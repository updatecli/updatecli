package gitbranch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type mockNativeGitHandler struct {
	gitgeneric.GitHandler
	branchRefs []gitgeneric.DatedBranch
}

func (m *mockNativeGitHandler) BranchRefs(workingDir string) ([]gitgeneric.DatedBranch, error) {
	return m.branchRefs, nil
}

// branchDaysAgo builds the latest commit date of a branch last updated n days ago.
func branchDaysAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * 24 * time.Hour)
}

func TestGitBranch_Source(t *testing.T) {
	branches := []gitgeneric.DatedBranch{
		{Name: "v1.0", Hash: "abc123", When: branchDaysAgo(30)},
		{Name: "v2.0", Hash: "def456", When: branchDaysAgo(10)},
		{Name: "v3.0", Hash: "ghi789", When: branchDaysAgo(1)},
	}

	tests := []struct {
		name        string
		spec        Spec
		wantValue   string
		wantSkipped bool
		wantErr     bool
	}{
		{
			name:      "no age filter returns the latest branch",
			spec:      Spec{},
			wantValue: "v3.0",
		},
		{
			name:      "the most recently updated branch is still in cooldown",
			spec:      Spec{Age: age.Spec{Minimum: "7d"}},
			wantValue: "v2.0",
		},
		{
			name:      "branches updated too long ago are discarded",
			spec:      Spec{Age: age.Spec{Maximum: "20d"}},
			wantValue: "v3.0",
		},
		{
			name:        "the source is skipped while every branch is still in cooldown",
			spec:        Spec{Age: age.Spec{Minimum: "60d"}},
			wantSkipped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := &GitBranch{
				spec:             tt.spec,
				versionFilter:    version.Filter{Kind: "latest", Pattern: "latest"},
				nativeGitHandler: &mockNativeGitHandler{branchRefs: branches},
			}

			gotResult := result.Source{}
			err := gb.Source(context.Background(), "/tmp/updatecli", &gotResult)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantSkipped {
				assert.Equal(t, result.SKIPPED, gotResult.Result)
				return
			}

			assert.Equal(t, tt.wantValue, gotResult.Information)
		})
	}
}
