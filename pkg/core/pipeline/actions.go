package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (p *Pipeline) RunActions() error {

	if len(p.Targets) == 0 {
		logrus.Debugln("no target found, skipping action")
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

		firstTargetID := relatedTargets[0]
		firstTargetSourceID := p.Targets[firstTargetID].Config.SourceID

		// If action.Title is not set then we try to guess it
		// based on the first target title
		if len(action.Title) == 0 {
			action.Title = p.Config.GetChangelogTitle(
				firstTargetID,
				p.Sources[firstTargetSourceID].Output,
			)
		}

		sourcesReport, err := p.Report.String("sources")
		if err != nil {
			return err
		}
		conditionsReport, err := p.Report.String("conditions")
		if err != nil {
			return err
		}
		targetsReport, err := p.Report.String("targets")
		if err != nil {
			return err
		}

		action.PipelineReport = fmt.Sprintf("%s\n%s\n%s\n",
			sourcesReport,
			conditionsReport,
			targetsReport,
		)

		if err != nil {
			return err
		}

		changelog := ""
		processedSourceIDs := []string{}

		failedTargetIDs, attentionTargetIDs, successTargetIDs, skippedTargetIDs := p.GetTargetsIDByResult(relatedTargets)

		// Ignoring failed targets
		if len(failedTargetIDs) > 0 {
			logrus.Errorf("%d target(s) (%s) failed for action %q", len(failedTargetIDs), strings.Join(failedTargetIDs, ","), id)
		}

		// Ignoring skipped targets
		if len(skippedTargetIDs) > 0 {
			return fmt.Errorf("%d target(s) (%s) skipped for action %q", len(skippedTargetIDs), strings.Join(skippedTargetIDs, ","), id)
		}

		// Ensure we don't add changelog from the same sourceID twice
		// Please note that the targets with both results (success and attention) need to be checked for changelog
		for _, targetID := range append(successTargetIDs, attentionTargetIDs...) {
			sourceID := p.Targets[targetID].Config.SourceID

			found := false
			for _, proccessedSourceID := range processedSourceIDs {
				if proccessedSourceID == sourceID {
					found = true
				}
			}
			if !found {
				changelog = changelog + p.Sources[sourceID].Changelog + "\n"
				processedSourceIDs = append(processedSourceIDs, sourceID)
			}
		}
		action.Changelog = changelog

		if p.Options.Target.DryRun {
			if len(attentionTargetIDs) > 0 {
				logrus.Infof("[Dry Run] An action of kind %q is expected.", action.Config.Kind)

				actionDebugOutput := fmt.Sprintf("The expected action would have the following informations:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
					action.Title,
					action.Changelog,
					action.PipelineReport)
				logrus.Debugf(strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))
			}

			actionOutput := fmt.Sprintf("The expected action would have the following informations:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
				action.Title,
				action.Changelog,
				action.PipelineReport)
			logrus.Debugf(strings.ReplaceAll(actionOutput, "\n", "\n\t|\t"))

			return nil
		}

		err = action.Handler.CreateAction(
			action.Title,
			action.Changelog,
			action.PipelineReport)

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
