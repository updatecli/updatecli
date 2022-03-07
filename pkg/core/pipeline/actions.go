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

	for id, act := range p.Actions {
		if _, ok := p.SCMs[act.Config.ScmID]; !ok {
			return fmt.Errorf("scm id %q couldn't be found", act.Config.ScmID)
		}

		inheritedSCM := p.SCMs[act.Config.ScmID]
		act.Scm = &inheritedSCM

		act = p.Actions[id]
		act.Config = p.Config.Spec.Actions[id]

		err := act.Update()
		if err != nil {
			return err
		}

		if p.SCMs[act.Config.ScmID].Config.Kind != act.Config.Kind {
			return fmt.Errorf("action of kind %q is not compatible with scm %q of kind %q",
				act.Config.Kind,
				act.Config.ScmID,
				p.SCMs[act.Config.ScmID].Config.Kind)
		}

		firstTargetID := act.Config.Targets[0]
		firstTargetSourceID := p.Targets[firstTargetID].Config.SourceID

		// If pr.Title is not set then we try to guess it
		// based on the first target title
		if len(act.Title) == 0 {
			act.Title = p.Config.GetChangelogTitle(
				firstTargetID,
				p.Sources[firstTargetSourceID].Result,
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

		act.PipelineReport = fmt.Sprintf("%s\n%s\n%s\n",
			sourcesReport,
			conditionsReport,
			targetsReport,
		)

		if err != nil {
			return err
		}

		changelog := ""
		processedSourceIDs := []string{}

		failedTargetIDs, attentionTargetIDs, successTargetIDs, skippedTargetIDs := p.GetTargetsIDByResult(act.Config.Targets)

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
		act.Changelog = changelog

		if p.Options.Target.DryRun {
			if len(attentionTargetIDs) > 0 {
				logrus.Infof("[Dry Run] A pull request with kind %q and title %q is expected.", act.Config.Kind, act.Title)

				actionDebugOutput := fmt.Sprintf("The expected pull request would have the following informations:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
					act.Title,
					act.Changelog,
					act.PipelineReport)
				logrus.Debugf(strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))
			}

			actionDebugOutput := fmt.Sprintf("The expected Pull Request would have the following informations:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
				act.Title,
				act.Changelog,
				act.PipelineReport)
			logrus.Debugf(strings.ReplaceAll(actionDebugOutput, "\n", "\n\t|\t"))

			return nil
		}

		err = act.Handler.Create(
			act.Title,
			act.Changelog,
			act.PipelineReport)

		if err != nil {
			return err
		}

		p.Actions[id] = act
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
