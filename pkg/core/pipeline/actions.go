package pipeline

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// CheckedPipelines is used to avoid checking the same action multiple times across different pipelines
	CheckedPipelines []string
)

// RunActions runs all actions defined in the configuration.
func (p *Pipeline) RunActions() error {

	if len(p.Actions) == 0 {
		logrus.Debugf("No action found for pipeline %q", p.Name)
		return nil
	}

	// Early return
	if len(p.Targets) == 0 {
		logrus.Debugf("No target found for pipeline %q, skipping depends on action", p.Name)
		return nil
	}

	// firstPipelineAction is used to display the pipeline name only once
	firstPipelineAction := true

	for id := range p.Actions {

		// Update pipeline before each action run
		if err := p.Update(); err != nil {
			logrus.Errorf("Pipeline %q failed to update: %s", p.Name, err.Error())
			continue
		}

		action := p.Actions[id]

		// action.Report.ID and action.Report.Title must be set
		// after actionTarget is set
		updateActionTitle := func() {
			action.Title = p.Config.Spec.Actions[id].Title
			// If an action spec doesn't have a title, then we use the one specified by the pipeline spec title
			if action.Title == "" {
				p.detectActionTitle(&action)
			}
		}

		showActionTitle := func() {
			updateActionTitle()
			if firstPipelineAction {
				logrus.Infof("\n%s", p.Name)
			}
			logrus.Infof("  => %s\n\n", action.Title)
			firstPipelineAction = false
		}

		relatedTargets, err := p.searchAssociatedTargetsID(id)
		if err != nil {
			logrus.Errorf("searching associated targets ID: %s", err.Error())
			continue
		}

		if len(relatedTargets) == 0 {
			logrus.Infof("No target found for action %q", id)
			continue
		}

		// alreadyCheckedAction is used to avoid checking the same action multiple times
		alreadyCheckedAction := p.Report.Targets[relatedTargets[0]].Scm.ID + p.ID

		if _, ok := p.SCMs[action.Config.ScmID]; !ok {
			return fmt.Errorf("scm id %q couldn't be found", action.Config.ScmID)
		}

		inheritedSCM := p.SCMs[action.Config.ScmID]
		action.Scm = &inheritedSCM

		action = p.Actions[id]
		action.Config = p.Config.Spec.Actions[id]

		if p.Report.Actions == nil {
			p.Report.Actions = make(map[string]*reports.Action)
		}
		p.Report.Actions[id] = &action.Report

		if err := action.Update(); err != nil {
			return err
		}

		failedTargetIDs, attentionTargetIDs, _, skippedTargetIDs := p.GetTargetsIDByResult(relatedTargets)

		// Ignoring failed targets
		if len(failedTargetIDs) > 0 {
			logrus.Errorf("%d target(s) (%s) failed for action %q", len(failedTargetIDs), strings.Join(failedTargetIDs, ","), id)
		}

		// Ignoring skipped targets
		if len(skippedTargetIDs) > 0 {
			logrus.Debugf("%d target(s) (%s) skipped for action %q", len(skippedTargetIDs), strings.Join(skippedTargetIDs, ","), id)
		}

		// If no target require attention while processing action in a attention state,
		// then we skip the action
		if len(attentionTargetIDs) == 0 {
			// If nothing changed within associated target, then we want to be sure that
			// we don't have an open pull request or an open issue before skipping the action
			// The goal is to identify pull request opened in previous Updatecli execution.
			if len(relatedTargets) > 0 {
				if !slices.Contains(CheckedPipelines, alreadyCheckedAction) {

					showActionTitle()

					err = action.Handler.CheckActionExist(&action.Report)
					if err != nil {
						logrus.Errorf("Action %q failed: %s", id, err.Error())
					}

					CheckedPipelines = append(CheckedPipelines, alreadyCheckedAction)
					if action.Report.Link == "" {
						logrus.Infof("No follow up action needed")
					}

					p.Report.Actions[id] = &action.Report
				}
			}
			continue
		}

		showActionTitle()

		// We try to identify if any of the related targets have a branch reset.
		// It it's the case then we set the action to have a branch reset
		// and remove the previous action description as it may be outdated.
		isBranchReset := false
		for _, t := range relatedTargets {
			if p.Targets[t].Result.Scm.BranchReset {
				isBranchReset = true
				break
			}
		}
		for _, t := range attentionTargetIDs {

			actionTarget := reports.ActionTarget{
				// Better for ID to use hash string
				ID:          fmt.Sprintf("%x", sha256.Sum256([]byte(t))),
				Title:       p.Targets[t].Config.Name,
				Description: p.Targets[t].Result.Description,
			}

			if len(p.Targets[t].Result.Changelogs) > 0 {

				for _, changelog := range p.Targets[t].Result.Changelogs {
					actionTarget.Changelogs = append(actionTarget.Changelogs, reports.ActionTargetChangelog{
						Title:       changelog.Title,
						Description: changelog.Body,
					})
				}

			}

			action.Report.Targets = append(action.Report.Targets, actionTarget)
		}

		pipelineName := p.Config.Spec.Name
		if pipelineName == "" && p.Config.Spec.Title != "" {
			pipelineName = p.Config.Spec.Name
		} else if pipelineName == "" && p.Name != "" {
			pipelineName = p.Name
		}

		action.Report.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(p.Name)))
		action.Report.Title = action.Title
		action.Report.PipelineTitle = pipelineName

		if !action.Config.DisablePipelineURL {
			action.Report.UpdatePipelineURL()
		}
		if isBranchReset {
			logrus.Warningf("Git branch reset detected, the action must reset previous action description")
		}

		if p.Options.Target.DryRun || !p.Options.Target.Push {
			if len(attentionTargetIDs) > 0 {
				logrus.Infof("[Dry Run] An action of kind %q is expected.", action.Config.Kind)

				actionDebugOutput := fmt.Sprintf("The expected action would have the following information:\n\n##Title:\n%s\n##Report:\n\n%s\n\n=====\n",
					action.Title,
					action.Report.String())
				logrus.Debugf("%s", strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))
			}

			actionOutput := fmt.Sprintf("The expected action would have the following information:\n\n##Title:\n%s\n\n\n##Report:\n\n%s\n\n=====\n",
				action.Title,
				action.Report.String())
			logrus.Debugf("%s", strings.ReplaceAll(actionOutput, "\n", "\n\t|\t"))

			action.Report.Description = actionOutput

			p.Report.Actions[id] = &action.Report

			return nil
		}

		err = action.Handler.CreateAction(&action.Report, isBranchReset)
		if err != nil {
			return err
		}

		isBranchReset = false

		p.Actions[id] = action
	}
	return nil
}

// RunCleanActions executes clean up operation which depends on the action plugin.
func (p *Pipeline) RunCleanActions() error {
	var errs []string

	// Early return
	if len(p.Targets) == 0 || len(p.Actions) == 0 {
		return nil
	}

	for _, action := range p.Actions {
		if !p.Options.Target.DryRun {
			if action.Handler != nil {
				// At least we try to clean existing pullrequest
				err := action.Handler.CleanAction(&action.Report)
				if err != nil {
					errs = append(errs, err.Error())
				}
			}
		}
	}

	if len(errs) > 0 {
		logrus.Errorf("Failed to clean up action(s): %s", strings.Join(errs, ", "))
		for i := range errs {
			logrus.Errorf("  - %s", errs[i])
		}
		return fmt.Errorf("failed to clean up action(s):\n\t%s", strings.Join(errs, "\n\t* "))
	}

	return nil
}

// GetTargetsIDByResult return a list of target ID per result type
func (p *Pipeline) GetTargetsIDByResult(targetIDs []string) (
	failedTargetsID, attentionTargetsID, successTargetsID, skippedTargetsID []string) {

	for _, targetID := range targetIDs {

		switch p.Report.Targets[targetID].Result {
		case result.FAILURE:
			failedTargetsID = append(failedTargetsID, targetID)

		case result.ATTENTION:
			attentionTargetsID = append(attentionTargetsID, targetID)

		case result.SKIPPED:
			skippedTargetsID = append(skippedTargetsID, targetID)

		case result.SUCCESS:
			successTargetsID = append(successTargetsID, targetID)
		}
	}
	return failedTargetsID, attentionTargetsID, successTargetsID, skippedTargetsID
}

// searchAssociatedTargetsID search for targets related to an action based on a scm configuration
func (p *Pipeline) searchAssociatedTargetsID(actionID string) ([]string, error) {

	scmid := p.Actions[actionID].Config.ScmID

	if len(scmid) == 0 {
		return []string{}, fmt.Errorf("scmid %q not found for the action id %q", scmid, actionID)
	}
	results := []string{}

	for id, target := range p.Targets {
		if target.Config.SCMID == scmid {
			results = append(results, id)
		}
	}
	return results, nil
}

// detectActionTitle set the action title based on the pipeline title or the target title
func (p *Pipeline) detectActionTitle(action *action.Action) {
	if action.Title != "" {
		return
	}

	if p.Config.Spec.Name != "" {
		action.Title = p.Config.Spec.Name
		return
	}

	// Title is deprecated and should be removed in the future
	if p.Config.Spec.Title != "" {
		action.Title = p.Config.Spec.Title
		return
	}

	if len(action.Report.Targets) == 0 {
		logrus.Debugf("We don't know how to name the action")
		return
	}

	// Search first validate action title based on target title
	// if none could be found then actionTitle keeps its default value
	for i := range action.Report.Targets {
		if action.Report.Targets[i].Title != "" {
			action.Title = action.Report.Targets[i].Title
			return
		}
	}
}
