package engine

import (
	"context"
	"errors"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
	"go.opentelemetry.io/otel/trace/noop"
)

// mockPushScm implements the scm methods used while pushing pending git changes.
type mockPushScm struct {
	scm.ScmHandler
	url           string
	workingBranch string
	targetBranch  string
	// remoteBranchUpToDate reports if the remote branch already contains the local commits
	remoteBranchUpToDate bool

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

// mockCloneScm implements the scm methods used while cloning git repositories.
type mockCloneScm struct {
	scm.ScmHandler
	directory string
	cloneErr  error
}

func (m *mockCloneScm) GetDirectory() string {
	return m.directory
}

func (m *mockCloneScm) Clone() (string, error) {
	return m.directory, m.cloneErr
}

func TestInitSCMFailedClone(t *testing.T) {
	failing := &pipeline.Pipeline{
		Name: "failing",
		SCMs: map[string]scm.Scm{
			"broken": {Handler: &mockCloneScm{directory: "/tmp/broken", cloneErr: errors.New("authentication required")}},
		},
	}
	working := &pipeline.Pipeline{
		Name: "working",
		SCMs: map[string]scm.Scm{
			"default": {Handler: &mockCloneScm{directory: "/tmp/default"}},
		},
	}

	e := Engine{Pipelines: []*pipeline.Pipeline{failing, working}}

	require.NoError(t, e.InitSCM())

	assert.EqualError(t, e.failedSCMError(failing), `scm "broken" failed to clone: authentication required`)
	assert.NoError(t, e.failedSCMError(working))
}

type recordingActionHandler struct {
	action.ActionHandler
	cleaned int
}

func (h *recordingActionHandler) CleanAction(ctx context.Context, report *reports.Action) error {
	h.cleaned++
	return nil
}

func TestRunActionsSkipsFailedClonePipeline(t *testing.T) {
	handler := &recordingActionHandler{}
	failing := &pipeline.Pipeline{
		Name:    "failing",
		SCMs:    map[string]scm.Scm{"broken": {Handler: &mockCloneScm{directory: "/tmp/broken", cloneErr: errors.New("authentication required")}}},
		Targets: map[string]target.Target{"t": {}},
		Actions: map[string]action.Action{"pr": {Handler: handler}},
	}

	e := Engine{Pipelines: []*pipeline.Pipeline{failing}}
	require.NoError(t, e.InitSCM())

	require.NoError(t, e.runActions(context.Background()))
	assert.Zero(t, handler.cleaned, "a pipeline whose scm failed to clone must not clean up or merge actions")
}

func TestRunReportsFailedClone(t *testing.T) {
	failing := &pipeline.Pipeline{
		Name: "failing",
		SCMs: map[string]scm.Scm{
			"broken": {Handler: &mockCloneScm{directory: "/tmp/broken", cloneErr: errors.New("authentication required")}},
		},
	}

	e := Engine{Pipelines: []*pipeline.Pipeline{failing}}
	e.Options.DisableUdashReport = true
	e.SetTracer(noop.NewTracerProvider().Tracer("test"))
	require.NoError(t, e.InitSCM())

	require.Error(t, e.Run(context.Background()))
	assert.Equal(t, result.FAILURE, failing.Report.Result)
	assert.Equal(t, `scm "broken" failed to clone: authentication required`, failing.Report.Err)
}
