package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (p *Pipeline) RunPullRequests() error {
	if len(p.PullRequests) > 0 {
		logrus.Infof("\n\n%s\n", strings.ToTitle("Pull Requests"))
		logrus.Infof("%s\n", strings.Repeat("=", len("PullRequests")+1))
	}

	for id, pr := range p.PullRequests {
		if _, ok := p.SCMs[pr.Config.ScmID]; !ok {
			return fmt.Errorf("scm id %q couldn't be found", pr.Config.ScmID)
		}

		inheritedSCM := p.SCMs[pr.Config.ScmID]
		pr.Scm = &inheritedSCM

		err := pr.Update()
		if err != nil {
			return err
		}

		if p.SCMs[pr.Config.ScmID].Config.Kind != pr.Config.Kind {
			return fmt.Errorf("pullrequest of kind %q is not compatible with scm %q of kind %q",
				pr.Config.Kind,
				pr.Config.ScmID,
				p.SCMs[pr.Config.ScmID].Config.Kind)
		}

		firstTargetID := pr.Config.Targets[0]
		firstTargetSourceID := p.Targets[firstTargetID].Config.SourceID

		// If pr.Title is not set then we try to guess it
		// based on the first target title
		if len(pr.Title) == 0 {
			pr.Title = p.Config.GetChangelogTitle(
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

		pr.PipelineReport = fmt.Sprintf("%s\n%s\n%s\n",
			sourcesReport,
			conditionsReport,
			targetsReport,
		)

		if err != nil {
			return err
		}

		changelog := ""
		processedSourceIDs := []string{}

		failedTargetIDs, attentionTargetIDs, successTargetIDs, skippedTargetIDs := p.GetTargetsIDByResult(pr.Config.Targets)

		// No target changed so nothing left to do
		if len(attentionTargetIDs) == 0 {
			logrus.Debugf("Nothing left do as no target related to pullrequest %q changed", id)
			return nil
		}

		// Ignoring failed targets
		if len(failedTargetIDs) > 0 {
			logrus.Debugf("%d target(s) (%s) failed for pullrequest %q", len(failedTargetIDs), strings.Join(failedTargetIDs, ","), id)
		}

		// Ignoring skipped targets
		if len(skippedTargetIDs) > 0 {
			return fmt.Errorf("%d  target(s) (%s) skipped for pullrequest %q", len(skippedTargetIDs), strings.Join(skippedTargetIDs, ","), id)
		}

		// Ensure we don't add changelog from the same sourceID twice.
		for _, targetID := range successTargetIDs {
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
		pr.Changelog = changelog

		if p.Options.Target.DryRun {
			pullRequestOutput := fmt.Sprintf("A Pull request of type %q, should be opened with following information:\n\n##Title:\n%s\n\n##Changelog:\n\n%s\n\n##Report:\n\n%s\n\n=====\n",
				pr.Config.Kind,
				pr.Title,
				pr.Changelog,
				pr.PipelineReport)
			logrus.Debugf(strings.ReplaceAll(pullRequestOutput, "\n", "\n\t|\t"))

			return nil
		}

		err = pr.Handler.CreatePullRequest(
			pr.Title,
			pr.Changelog,
			pr.PipelineReport)

		if err != nil {
			return err
		}

		p.PullRequests[id] = pr
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
