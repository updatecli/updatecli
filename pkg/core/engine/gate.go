package engine

import (
	"context"
	"fmt"
	"slices"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// waitingTrack returns why a pipeline generated from autodiscovery "tracks" must wait,
// or an empty string when it can run.
//
// A pipeline waits while a pipeline for the same dependency, with an earlier track,
// has a pending update: either it changed something during this run,
// or its pull request from a previous run is still open.
// pending caches, for the whole run, the lookups that succeeded and the failures already warned about.
func (e *Engine) waitingTrack(ctx context.Context, p *pipeline.Pipeline, pending map[string]bool) string {
	if p.Gate == nil || p.Gate.Rank == 0 {
		return ""
	}

	for _, earlier := range e.Pipelines {
		if earlier == p || earlier.Gate == nil {
			continue
		}

		if earlier.Gate.Key != p.Gate.Key || earlier.Gate.Rank >= p.Gate.Rank {
			continue
		}

		if hasPendingUpdate(ctx, earlier, pending) {
			return fmt.Sprintf("waiting for the pending update of track %q", earlier.Gate.Track)
		}
	}

	return ""
}

// hasPendingUpdate reports whether a pipeline changed something during this run,
// or still has an update waiting to be merged.
//
// With actions, such as a pull request, the update waits while the action is still open:
// a working branch left behind after its pull request is merged or closed doesn't count.
// Without actions, or when none of them could be queried, the working branch is the update,
// so the update waits while the branch exists.
//
// A pipeline fails as soon as one target fails, while its other targets may still open a pull request,
// so a target that changed something counts as pending whatever the pipeline result.
// A failed pipeline that changed nothing doesn't block, otherwise a failure repeated on every run,
// such as a version filter matching no version, would block later tracks for good.
// A failed lookup counts as pending, so a later track never goes ahead on a lookup error.
func hasPendingUpdate(ctx context.Context, p *pipeline.Pipeline, pending map[string]bool) bool {
	if p.Report.Result == result.ATTENTION || hasChangedTarget(p) {
		return true
	}

	var waiting bool
	if open, checked := hasOpenAction(ctx, p, pending); checked {
		waiting = open
	} else {
		waiting = hasRemoteWorkingBranch(p, pending)
	}

	// Warn once per run, as every later track checks this pipeline.
	if !waiting && p.Report.Result == result.FAILURE {
		warnedKey := "failure#" + p.ID
		if !pending[warnedKey] {
			pending[warnedKey] = true
			log.Warningf("pipeline %q, for track %q, failed without changing anything: later tracks don't wait for it",
				p.Name, p.Gate.Track)
		}
	}

	return waiting
}

// hasChangedTarget reports whether one of the pipeline targets changed something during this run.
func hasChangedTarget(p *pipeline.Pipeline) bool {
	for _, t := range p.Targets {
		if t.Result != nil && (t.Result.Result == result.ATTENTION || t.Result.Changed) {
			return true
		}
	}
	return false
}

// hasOpenAction reports whether one of the pipeline actions, such as a pull request, is still open.
// checked is false when the pipeline has no action that could be queried.
func hasOpenAction(ctx context.Context, p *pipeline.Pipeline, pending map[string]bool) (open, checked bool) {
	ids := make([]string, 0, len(p.Actions))
	for id := range p.Actions {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	for _, id := range ids {
		a := p.Actions[id]
		if a.Handler == nil {
			continue
		}
		checked = true

		key := "action#" + p.ID + "#" + id
		isOpen, ok := pending[key]
		if !ok {
			report := reports.Action{}
			if err := a.Handler.CheckActionExist(ctx, &report); err != nil {
				log.Warningf("checking if action %q of pipeline %q is still open: %s", id, p.Name, err)
				return true, true
			}

			isOpen = report.Link != ""
			pending[key] = isOpen
		}

		if isOpen {
			return true, true
		}
	}

	return false, checked
}

// hasRemoteWorkingBranch reports whether the working branch of one of the pipeline targets exists on its remote.
func hasRemoteWorkingBranch(p *pipeline.Pipeline, pending map[string]bool) bool {
	for _, target := range p.Targets {
		if target.Scm == nil || *target.Scm == nil {
			continue
		}

		s := *target.Scm
		_, workingBranch, targetBranch := s.GetBranches()
		if workingBranch == targetBranch {
			continue
		}

		key := "branch#" + s.GetURL() + "#" + workingBranch
		exists, ok := pending[key]
		if !ok {
			var err error
			exists, err = s.IsRemoteWorkingBranchExist()
			if err != nil {
				log.Warningf("checking if remote branch %q exists: %s", workingBranch, err)
				return true
			}
			pending[key] = exists
		}

		if exists {
			return true
		}
	}

	return false
}

// skipWaitingPipeline leaves a waiting pipeline and its targets skipped, with the reason in each target report.
// Its targets don't run, so its working branch is neither pushed nor deleted,
// and its actions are skipped, so an open pull request is neither updated nor merged.
func skipWaitingPipeline(p *pipeline.Pipeline, reason string) {
	log.Infof("%s Skipping pipeline %q: %s", result.SKIPPED, p.Name, reason)

	p.Report.Result = result.SKIPPED
	if p.Gate != nil {
		p.Gate.Waiting = true
	}
	for id := range p.Targets {
		if p.Targets[id].Result != nil {
			p.Targets[id].Result.Description = reason
		}
	}
}
