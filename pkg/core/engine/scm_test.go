package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// mockPushScm implements the scm methods used while pushing pending git changes.
type mockPushScm struct {
	scm.ScmHandler
	url           string
	workingBranch string
	targetBranch  string
	// remoteBranchUpToDate reports if the remote branch already contains the local commits
	remoteBranchUpToDate bool

	// pushErrors holds the error returned by each successive Push call, if any.
	// A nil entry makes the corresponding Push call succeed.
	pushErrors []error

	pushCounter               int
	cleanWorkingBranchCounter int
}

func (m *mockPushScm) GetURL() string {
	return m.url
}

func (m *mockPushScm) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	return m.targetBranch, m.workingBranch, m.targetBranch
}

func (m *mockPushScm) CleanWorkingBranch() (bool, error) {
	m.cleanWorkingBranchCounter++
	return true, nil
}

func (m *mockPushScm) IsRemoteBranchUpToDate() (bool, error) {
	return m.remoteBranchUpToDate, nil
}

func (m *mockPushScm) Checkout() error {
	return nil
}

func (m *mockPushScm) Push() (bool, error) {
	m.pushCounter++
	if m.pushCounter <= len(m.pushErrors) {
		if err := m.pushErrors[m.pushCounter-1]; err != nil {
			return false, err
		}
	}
	return true, nil
}

func TestPushSCMCommits(t *testing.T) {
	testdata := []struct {
		name string
		// targetResult is the result of the only target attached to the scm
		targetResult string
		// expectedPushCounter reports how many times the working branch was published
		expectedPushCounter int
	}{
		{
			/*
				All targets ran so Updatecli knows the content that the working branch
				must have, such as a working branch reset because the change is not
				needed anymore.
			*/
			name:                "local branch of a fully executed pipeline is published",
			targetResult:        result.SUCCESS,
			expectedPushCounter: 1,
		},
		{
			/*
				The target didn't run, for example because its source failed, so the
				working branch doesn't hold the expected content.
				Publishing it would remove the changes pushed by a previous execution and
				would then close the associated pull request.
			*/
			name:                "local branch is not published when a target didn't run",
			targetResult:        result.SKIPPED,
			expectedPushCounter: 0,
		},
		{
			name:                "local branch is not published when a target failed",
			targetResult:        result.FAILURE,
			expectedPushCounter: 0,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			mockScm := mockPushScm{
				url:                  "https://github.com/updatecli/updatecli.git",
				workingBranch:        "updatecli_main",
				targetBranch:         "main",
				remoteBranchUpToDate: false,
			}

			var scmHandler scm.ScmHandler = &mockScm

			e := Engine{
				Pipelines: []*pipeline.Pipeline{
					{
						Name: "test",
						Targets: map[string]target.Target{
							"default": {
								Scm:    &scmHandler,
								Result: &result.Target{Result: tt.targetResult},
							},
						},
					},
				},
			}

			gotErr := e.pushSCMCommits()

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedPushCounter, mockScm.pushCounter)
		})
	}
}

func TestPruneSCMBranches(t *testing.T) {
	testdata := []struct {
		name string
		// targetResult is the result of the only target attached to the scm
		targetResult string
		// expectedCleanWorkingBranchCounter reports how many times the working branch cleanup ran
		expectedCleanWorkingBranchCounter int
	}{
		{
			name:                              "working branch of a fully executed pipeline can be cleaned",
			targetResult:                      result.SUCCESS,
			expectedCleanWorkingBranchCounter: 1,
		},
		{
			/*
				Deleting the working branch closes the pull request associated with it, so
				it must not happen based on an execution which didn't run its target(s).
			*/
			name:                              "working branch is not cleaned when a target didn't run",
			targetResult:                      result.SKIPPED,
			expectedCleanWorkingBranchCounter: 0,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			mockScm := mockPushScm{
				url:           "https://github.com/updatecli/updatecli.git",
				workingBranch: "updatecli_main",
				targetBranch:  "main",
			}

			var scmHandler scm.ScmHandler = &mockScm

			e := Engine{
				Pipelines: []*pipeline.Pipeline{
					{
						Name: "test",
						Targets: map[string]target.Target{
							"default": {
								Scm:    &scmHandler,
								Result: &result.Target{Result: tt.targetResult},
							},
						},
					},
				},
			}

			gotErr := e.pruneSCMBranches()

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedCleanWorkingBranchCounter, mockScm.cleanWorkingBranchCounter)
		})
	}
}

// TestPushSCMCommitsRecoversFromPushErrors checks that a working branch which
// ends up pushed is not reported as a failure because an earlier push attempt
// of the same branch errored. See https://github.com/updatecli/updatecli/issues/10825
func TestPushSCMCommitsRecoversFromPushErrors(t *testing.T) {
	refLockErr := errors.New("cannot lock ref 'refs/heads/updatecli_main': reference already exists")

	testdata := []struct {
		name string
		// pipelines is the number of pipelines sharing the same SCM working branch
		pipelines            int
		pushErrors           []error
		remoteBranchUpToDate bool
		expectedError        bool
		expectedPushCounter  int
	}{
		{
			name:                 "an earlier push error is cleared when a later target pushes the branch",
			pipelines:            2,
			pushErrors:           []error{refLockErr, nil},
			remoteBranchUpToDate: true,
			expectedError:        false,
			expectedPushCounter:  2,
		},
		{
			name:                 "an earlier push error is cleared when the remote branch is up to date",
			pipelines:            1,
			pushErrors:           []error{refLockErr},
			remoteBranchUpToDate: true,
			expectedError:        false,
			expectedPushCounter:  1,
		},
		{
			name:                 "push errors are kept when the branch cannot be published",
			pipelines:            1,
			pushErrors:           []error{refLockErr, refLockErr},
			remoteBranchUpToDate: false,
			expectedError:        true,
			expectedPushCounter:  2,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			mockScm := mockPushScm{
				url:                  "https://github.com/updatecli/updatecli.git",
				workingBranch:        "updatecli_main",
				targetBranch:         "main",
				remoteBranchUpToDate: tt.remoteBranchUpToDate,
				pushErrors:           tt.pushErrors,
			}

			var scmHandler scm.ScmHandler = &mockScm

			pipelines := make([]*pipeline.Pipeline, 0, tt.pipelines)
			targetResults := make([]*result.Target, 0, tt.pipelines)
			for i := 0; i < tt.pipelines; i++ {
				targetResult := &result.Target{Result: result.SUCCESS}
				targetResults = append(targetResults, targetResult)
				pipelines = append(pipelines, &pipeline.Pipeline{
					Name: "test",
					Targets: map[string]target.Target{
						"default": {
							Scm:    &scmHandler,
							ToPush: true,
							Result: targetResult,
						},
					},
				})
			}

			e := Engine{Pipelines: pipelines}

			gotErr := e.pushSCMCommits()

			if tt.expectedError {
				require.Error(t, gotErr)
				for _, targetResult := range targetResults {
					// A branch which is not published must keep its targets failed,
					// otherwise actions would run for changes that were not pushed.
					assert.Equal(t, result.FAILURE, targetResult.Result)
					assert.Equal(t, "failed to push commits", targetResult.Description)
				}
			} else {
				require.NoError(t, gotErr)
				for _, targetResult := range targetResults {
					// Once the branch is published the previous target result must be
					// restored, otherwise the action creating the pull request is
					// skipped as it only runs for targets requiring attention.
					assert.Equal(t, result.SUCCESS, targetResult.Result)
					assert.Empty(t, targetResult.Description)
				}
			}
			assert.Equal(t, tt.expectedPushCounter, mockScm.pushCounter)
		})
	}
}
