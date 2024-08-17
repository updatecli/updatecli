package pipeline

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunActions runs all actions defined in the configuration.
func (p *Pipeline) RunActions() error {

	// Early return
	if len(p.Targets) == 0 || len(p.Actions) == 0 {
		return nil
	}

	for id := range p.Actions {

		action := p.Actions[id]

		relatedTargets, err := p.searchAssociatedTargetsID(id)
		if err != nil {
			logrus.Errorf(err.Error())
			continue
		}

		// Update pipeline before each condition run
		if err = p.Update(); err != nil {
			logrus.Errorf(err.Error())
			continue
		}

		if _, ok := p.SCMs[action.Config.ScmID]; !ok {
			return fmt.Errorf("scm id %q couldn't be found", action.Config.ScmID)
		}

		inheritedSCM := p.SCMs[action.Config.ScmID]
		action.Scm = &inheritedSCM

		action = p.Actions[id]
		action.Config = p.Config.Spec.Actions[id]

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
			continue
		}

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

			if p.Sources[p.Targets[t].Config.SourceID].Changelog != "" {
				actionTarget.Changelogs = append(actionTarget.Changelogs, reports.ActionTargetChangelog{
					Title:       p.Sources[p.Targets[t].Config.SourceID].Output,
					Description: p.Sources[p.Targets[t].Config.SourceID].Changelog,
				})
			}

			action.Report.Targets = append(action.Report.Targets, actionTarget)
		}

		logrus.Infof("\n%s - %s", p.Name, id)
		logrus.Infof("%s\n\n", strings.Repeat("-", len(p.Name)+len(id)+3))

		// Must action.Report.ID and action.Report.Title must be set
		// after actionTarget is set
		actionTitle := action.Title
		// If an action spec doesn't have a title, then we use the one specified by the pipeline spec title
		if actionTitle == "" && p.Config.Spec.Name != "" {
			actionTitle = p.Config.Spec.Name
		} else if actionTitle == "" && p.Config.Spec.Title != "" {
			// The field "Title" is probably useless and need to be refactor in a later iteration
			actionTitle = p.Config.Spec.Title
		} else {
			actionTitle = getActionTitle(action)
		}

		pipelineName := p.Config.Spec.Name
		if pipelineName == "" && p.Config.Spec.Title != "" {
			pipelineName = p.Config.Spec.Name
		} else if pipelineName == "" && p.Name != "" {
			pipelineName = p.Name
		}

		action.Report.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(p.Name)))
		action.Report.Title = actionTitle
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
					actionTitle,
					action.Report.String())
				logrus.Debugf(strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))
			}

			actionOutput := fmt.Sprintf("The expected action would have the following information:\n\n##Title:\n%s\n\n\n##Report:\n\n%s\n\n=====\n",
				actionTitle,
				action.Report.String())
			logrus.Debugf(strings.ReplaceAll(actionOutput, "\n", "\n\t|\t"))

			return nil
		}

		err = action.Handler.CreateAction(action.Report, isBranchReset)
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
			if (action.Handler != nil) {
				// At least we try to clean existing pullrequest
				err := action.Handler.CleanAction(action.Report)
				if err != nil {
					errs = append(errs, err.Error())
				}
			}
		}
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

func getActionTitle(action action.Action) string {
	switch action.Title != "" {
	case true:
		return action.Title
	case false:
		// Search first validate action title based on target title
		// if none could be found then actionTitle keeps its default value
		for i := range action.Report.Targets {
			if action.Report.Targets[i].Title != "" {
				return action.Report.Targets[i].Title

			}
		}
	}
	return "No action title could be found"
}
