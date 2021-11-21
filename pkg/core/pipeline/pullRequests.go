package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func (p *Pipeline) RunPullRequests() error {
	if len(p.PullRequests) > 0 {
		logrus.Infof("\n\n%s\n", strings.ToTitle("Pull Requests"))
		logrus.Infof("%s\n", strings.Repeat("=", len("PullRequests")+1))
	}

	for _, pr := range p.PullRequests {

		firstDependsOnTargetID := pr.Config.DependsOnTargets[0]
		firstDependsOntTargetSourceID := p.Targets[firstDependsOnTargetID].Config.SourceID

		// A new pullRequest title is based on the first target title
		pr.Title = p.Config.GetChangelogTitle(
			firstDependsOnTargetID,
			p.Sources[firstDependsOntTargetSourceID].Result,
		)

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
		// Ensure we don't add changelog from the same sourceID twice.
		for _, targetID := range pr.Config.DependsOnTargets {
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
			logrus.Debugf("A Pull request of type %q, should be open with following information:\n=====\nTitle:\t%s\nChangelog:\n%s\nReport:\n%s\n\n=====\n",
				pr.Config.Kind,
				pr.Title,
				pr.Changelog,
				pr.PipelineReport)

			return nil

		}

		err = pr.PullRequest.CreatePullRequest(
			pr.Title,
			pr.Changelog,
			pr.PipelineReport)

		if err != nil {
			return err
		}
	}
	return nil
}
