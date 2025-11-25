package mergerequest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRemoteBranchExist(t *testing.T) {

	testdata := []struct {
		name           string
		spec           Spec
		expectedResult bool
	}{
		{
			name: "Existing branch",
			spec: Spec{
				Owner:        "olblak",
				Repository:   "updatecli",
				SourceBranch: "main",
				TargetBranch: "main",
			},
			expectedResult: true,
		},
		{
			name: "Existing branch",
			spec: Spec{
				Owner:        "olblak",
				Repository:   "updatecli",
				SourceBranch: "donotexist",
				TargetBranch: "donotexist",
			},
			expectedResult: false,
		},
	}

	for _, td := range testdata {
		t.Run(td.name, func(t *testing.T) {
			gitlab, err := New(td.spec, nil)
			if err != nil {
				t.Fatalf("failed to create Gitlab instance: %v", err)
			}

			gotResult, err := gitlab.isRemoteBranchesExist()
			require.NoError(t, err)

			require.Equal(t, td.expectedResult, gotResult)

		})
	}
}
func TestFindExistingMR_Table(t *testing.T) {
	testdata := []struct {
		name string
		spec Spec
	}{
		{
			name: "NoMergeRequest",
			spec: Spec{
				Owner:        "olblak",
				Repository:   "updatecli",
				SourceBranch: "donotexist",
				TargetBranch: "donotexist",
			},
		},
		{
			name: "NoOpenedMergeRequestBetweenMainBranches",
			spec: Spec{
				Owner:        "olblak",
				Repository:   "updatecli",
				SourceBranch: "main",
				TargetBranch: "main",
			},
		},
	}

	for _, td := range testdata {
		tc := td
		t.Run(tc.name, func(t *testing.T) {
			gitlab, err := New(tc.spec, nil)
			if err != nil {
				t.Fatalf("failed to create Gitlab instance: %v", err)
			}

			mr, err := gitlab.findExistingMR()
			require.NoError(t, err)
			require.Nil(t, mr)
		})
	}
}
