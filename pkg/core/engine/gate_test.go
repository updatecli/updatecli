package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// mockGateScm implements the scm methods used to find a pending update.
type mockGateScm struct {
	scm.ScmHandler
	workingBranch string
	exists        bool
	err           error
	calls         int
}

func (m *mockGateScm) GetURL() string {
	return "https://example.com/org/repo.git"
}

func (m *mockGateScm) GetBranches() (sourceBranch, workingBranch, targetBranch string) {
	return "main", m.workingBranch, "main"
}

func (m *mockGateScm) IsRemoteWorkingBranchExist() (bool, error) {
	m.calls++
	return m.exists, m.err
}

// mockGateAction implements the action method used to find an open pull request.
type mockGateAction struct {
	action.ActionHandler
	link  string
	err   error
	calls int
}

func (m *mockGateAction) CheckActionExist(_ context.Context, report *reports.Action) error {
	m.calls++
	report.Link = m.link
	return m.err
}

// withAction adds an action, such as a pull request, to a gated pipeline.
func withAction(p *pipeline.Pipeline, handler action.ActionHandler) *pipeline.Pipeline {
	p.ID = p.Name
	p.Actions = map[string]action.Action{"default": {Handler: handler}}
	return p
}

func gatedPipeline(name, key string, rank int, filter, reportResult string, s scm.ScmHandler) *pipeline.Pipeline {
	p := &pipeline.Pipeline{
		Name: name,
		Gate: &pipeline.Gate{Key: key, Rank: rank, Track: filter},
		Targets: map[string]target.Target{
			"default": {Result: &result.Target{Result: result.SKIPPED}},
		},
	}
	p.Report.Result = reportResult

	if s != nil {
		t := p.Targets["default"]
		t.Scm = &s
		p.Targets["default"] = t
	}

	return p
}

func TestWaitingTrack(t *testing.T) {
	tests := []struct {
		name        string
		earlier     func() *pipeline.Pipeline
		laterKey    string
		wantWaiting bool
	}{
		{
			name: "earlier track changed something during this run",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.ATTENTION, nil)
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "earlier track still has a pull request from a previous run",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor", exists: true})
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "earlier track is up to date",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor"})
			},
			laterKey:    "k",
			wantWaiting: false,
		},
		{
			name: "earlier track without working branch",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "main", exists: true})
			},
			laterKey:    "k",
			wantWaiting: false,
		},
		{
			name: "remote lookup failure counts as pending",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor", err: errors.New("boom")})
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "earlier track still has an open pull request",
			earlier: func() *pipeline.Pipeline {
				return withAction(
					gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor"}),
					&mockGateAction{link: "https://example.com/org/repo/pull/1"},
				)
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "merged pull request with a leftover working branch doesn't block",
			earlier: func() *pipeline.Pipeline {
				return withAction(
					gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor", exists: true}),
					&mockGateAction{},
				)
			},
			laterKey:    "k",
			wantWaiting: false,
		},
		{
			name: "pull request lookup failure counts as pending",
			earlier: func() *pipeline.Pipeline {
				return withAction(
					gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, nil),
					&mockGateAction{err: errors.New("rate limited")},
				)
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "earlier track failed but another target changed something",
			earlier: func() *pipeline.Pipeline {
				p := gatedPipeline("minor", "k", 0, "minor", result.FAILURE, nil)
				p.Targets["changed"] = target.Target{Result: &result.Target{Result: result.ATTENTION, Changed: true}}
				return p
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "earlier track failed without changing anything",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "k", 0, "minor", result.FAILURE, nil)
			},
			laterKey:    "k",
			wantWaiting: false,
		},
		{
			name: "action without handler falls back to the working branch",
			earlier: func() *pipeline.Pipeline {
				return withAction(
					gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, &mockGateScm{workingBranch: "updatecli_main_minor", exists: true}),
					nil,
				)
			},
			laterKey:    "k",
			wantWaiting: true,
		},
		{
			name: "another dependency doesn't block",
			earlier: func() *pipeline.Pipeline {
				return gatedPipeline("minor", "other", 0, "minor", result.ATTENTION, nil)
			},
			laterKey:    "k",
			wantWaiting: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			earlier := tt.earlier()
			later := gatedPipeline("major", tt.laterKey, 1, "major", result.SKIPPED, nil)

			e := Engine{Pipelines: []*pipeline.Pipeline{earlier, later}}

			reason := e.waitingTrack(context.Background(), later, map[string]bool{})
			if tt.wantWaiting {
				assert.Contains(t, reason, `track "minor"`)
			} else {
				assert.Empty(t, reason)
			}

			// The first track never waits.
			assert.Empty(t, e.waitingTrack(context.Background(), earlier, map[string]bool{}))
		})
	}
}

func TestWaitingTrackNoWait(t *testing.T) {
	minor := gatedPipeline("minor", "k", 0, "minor", result.ATTENTION, nil)
	lts := gatedPipeline("lts", "k", 1, "lts", result.ATTENTION, nil)
	lts.Gate.NoWait = true
	major := gatedPipeline("major", "k", 2, "major", result.SKIPPED, nil)

	e := Engine{Pipelines: []*pipeline.Pipeline{minor, lts, major}}

	// A track that doesn't wait runs while an earlier track is pending.
	assert.Empty(t, e.waitingTrack(context.Background(), lts, map[string]bool{}))

	// Later tracks still wait for it.
	e.Pipelines = []*pipeline.Pipeline{lts, major}
	assert.Contains(t, e.waitingTrack(context.Background(), major, map[string]bool{}), `track "lts"`)
}

func TestWaitingTrackCachesRemoteLookups(t *testing.T) {
	s := &mockGateScm{workingBranch: "updatecli_main_minor"}
	earlier := gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, s)
	later := gatedPipeline("major", "k", 1, "major", result.SKIPPED, nil)
	latest := gatedPipeline("next", "k", 2, "next", result.SKIPPED, nil)

	e := Engine{Pipelines: []*pipeline.Pipeline{earlier, later, latest}}
	cache := map[string]bool{}

	assert.Empty(t, e.waitingTrack(context.Background(), later, cache))
	assert.Empty(t, e.waitingTrack(context.Background(), latest, cache))
	assert.Equal(t, 1, s.calls)
}

func TestWaitingTrackCachesActionLookups(t *testing.T) {
	a := &mockGateAction{}
	earlier := withAction(gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, nil), a)
	later := gatedPipeline("major", "k", 1, "major", result.SKIPPED, nil)
	latest := gatedPipeline("next", "k", 2, "next", result.SKIPPED, nil)

	e := Engine{Pipelines: []*pipeline.Pipeline{earlier, later, latest}}
	cache := map[string]bool{}

	assert.Empty(t, e.waitingTrack(context.Background(), later, cache))
	assert.Empty(t, e.waitingTrack(context.Background(), latest, cache))
	assert.Equal(t, 1, a.calls)
}

func TestWaitingTrackRetriesFailedLookups(t *testing.T) {
	s := &mockGateScm{workingBranch: "updatecli_main_minor", err: errors.New("boom")}
	earlier := gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, s)
	later := gatedPipeline("major", "k", 1, "major", result.SKIPPED, nil)
	latest := gatedPipeline("next", "k", 2, "next", result.SKIPPED, nil)

	e := Engine{Pipelines: []*pipeline.Pipeline{earlier, later, latest}}
	cache := map[string]bool{}

	assert.NotEmpty(t, e.waitingTrack(context.Background(), later, cache))
	assert.NotEmpty(t, e.waitingTrack(context.Background(), latest, cache))
	assert.Equal(t, 2, s.calls)
}

func TestSkipWaitingPipeline(t *testing.T) {
	p := gatedPipeline("major", "k", 1, "major", result.SUCCESS, nil)

	skipWaitingPipeline(p, "waiting")

	assert.Equal(t, result.SKIPPED, p.Report.Result)
	assert.True(t, p.IsWaiting())
	assert.Equal(t, result.SKIPPED, p.Targets["default"].Result.Result)
	assert.Equal(t, "waiting", p.Targets["default"].Result.Description)
}

func TestWaitingPipelineSkipsActions(t *testing.T) {
	waiting := gatedPipeline("major", "k", 1, "major", result.SUCCESS, nil)
	running := gatedPipeline("minor", "k", 0, "minor", result.SUCCESS, nil)
	e := Engine{Pipelines: []*pipeline.Pipeline{running, waiting}}

	assert.False(t, e.skipsActions(waiting))

	skipWaitingPipeline(waiting, "waiting")

	assert.True(t, e.skipsActions(waiting))
	assert.False(t, e.skipsActions(running))
}
