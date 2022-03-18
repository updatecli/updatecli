package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (p *Pipeline) RunActions() error {
	if len(p.Actions) > 0 {
		logrus.Infof("\n\n%s\n", strings.ToTitle("Actions"))
		logrus.Infof("%s\n\n", strings.Repeat("=", len("Actions")+1))
	}

	for id, action := range p.Actions {
		if _, ok := p.SCMs[action.Config.ScmID]; !ok {
			return fmt.Errorf("scm id %q couldn't be found", action.Config.ScmID)
		}

		inheritedSCM := p.SCMs[action.Config.ScmID]
		action.Scm = &inheritedSCM

		action = p.Actions[id]
		action.Config = p.Config.Spec.Actions[id]

		// Update configuration (case of templating used within YAML manifest)
		err := action.Update()
		if err != nil {
			return err
		}

		firstTargetID := action.Config.Targets[0]
		firstTargetSourceID := p.Targets[firstTargetID].Config.SourceID

		// If pr.Title is not set then we try to guess it
		// based on the first target title
		if len(action.Title) == 0 {
			action.Title = p.Config.GetChangelogTitle(
				firstTargetID,
				p.Sources[firstTargetSourceID].Result,
			)
		}

		// Execute specific "action" procedure on each associated target
		if !p.Options.DryRun {
			// Creates the temporary branch once for all
			if err = action.Scm.Handler.Checkout(); err != nil {
				return err
			}

			// for _, targetID := range action.Config.Targets {
			// 	err = action.Handler.RunTarget(p.Targets[targetID], p.Options)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
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

		changelog := ""
		processedSourceIDs := []string{}

		failedTargetIDs, attentionTargetIDs, successTargetIDs, skippedTargetIDs := p.GetTargetsIDByResult(action.Config.Targets)

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

		// Execute the action if there is at least one changed target
		if len(attentionTargetIDs) > 0 {
			err = action.Handler.Run(action.Title, action.Changelog, action.PipelineReport)
			if err != nil {
				return err
			}
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
