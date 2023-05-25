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

func (p *Pipeline) RunActions() error {

	if len(p.Targets) == 0 {
		logrus.Debugln("no target found, skipping action")
		return nil
	}

	if len(p.Actions) == 0 {
		logrus.Debugln("no action found, skipping")
		return nil
	}

	if len(p.Actions) > 0 {
		logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
		logrus.Infof("%s\n\n", strings.Repeat("=", len("Actions")+1))
	}

	for id, action := range p.Actions {
		relatedTargets, err := p.SearchAssociatedTargetsID(id)
		if err != nil {
			logrus.Errorf(err.Error())
			continue
		}

		// Update pipeline before each condition run
		err = p.Update()
		if err != nil {
			logrus.Errorf(err.Error())
			continue
		}

		if err != nil {
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

		err = action.Update()
		if err != nil {
			return err
		}

		failedTargetIDs, attentionTargetIDs, _, skippedTargetIDs := p.GetTargetsIDByResult(relatedTargets)

		/*
			Better for ID to use hash string
			By having a actionID that combine both the pipelineID and the actionID, avoid collision
			when two different pipeline open the same pullrequest based on the same action title
		*/

		for _, t := range relatedTargets {
			// We only care about target that have changed something
			if !p.Targets[t].Result.Changed {
				continue
			}

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

		// No need to execute the action if no target require attention
		if len(action.Report.Targets) == 0 {
			continue
		}

		// Must action.Report.ID and action.Report.Title must be set after actionTarget are set
		actionTitle := action.Title
		// If an action spec do not have a tittle, then we use the one specified by the pipeline spec title
		if actionTitle == "" && p.Config.Spec.Name != "" {
			actionTitle = p.Config.Spec.Name
		} else if actionTitle == "" && p.Config.Spec.Title != "" {
			// The field "Title" is probably useless and need to be refactor in a later iteration
			actionTitle = p.Config.Spec.Title
		} else {
			actionTitle = getActionTitle(action)
		}

		pipelineTitle := p.Config.Spec.Name
		if pipelineTitle == "" && p.Config.Spec.Title != "" {
			pipelineTitle = p.Config.Spec.Name
		} else if pipelineTitle == "" && p.Name != "" {
			pipelineTitle = p.Name
		}

		action.Report.ID = fmt.Sprintf("%x", sha256.Sum256([]byte(p.Title+actionTitle)))
		action.Report.Title = actionTitle
		action.Report.PipelineTitle = pipelineTitle

		// Ignoring failed targets
		if len(failedTargetIDs) > 0 {
			logrus.Errorf("%d target(s) (%s) failed for action %q", len(failedTargetIDs), strings.Join(failedTargetIDs, ","), id)
		}

		// Ignoring skipped targets
		if len(skippedTargetIDs) > 0 {
			return fmt.Errorf("%d target(s) (%s) skipped for action %q", len(skippedTargetIDs), strings.Join(skippedTargetIDs, ","), id)
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

		err = action.Handler.CreateAction(action.Report)

		if err != nil {
			return err
		}

		p.Actions[id] = action
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

// SearchAssociatedTargetsID search for targets related to an action based on a scm configuration
func (p *Pipeline) SearchAssociatedTargetsID(actionID string) ([]string, error) {

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
