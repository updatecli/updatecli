package pipeline

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrRunTargets is return when at least one error happened during targets execution
	ErrRunTargets error = errors.New("something went wrong during target execution")
)

func (p *Pipeline) updateTarget(id, result string) {

	target := p.Targets[id]
	target.Result.Result = result
	p.Targets[id] = target
	p.Report.Targets[id] = &target.Result
}

// RunTarget run a target by id
func (p *Pipeline) RunTarget(id string, sourceIds []string) (r string, changed bool, err error) {
	target := p.Targets[id]
	target.Config = p.Config.Spec.Targets[id]
	// Ensure the result named contains the up to date target name after templating
	target.Result.Name = target.Config.Name
	target.Result.DryRun = target.DryRun

	err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)
	if err != nil {
		p.Report.Result = result.FAILURE
		target.Result.Result = result.FAILURE
		err = fmt.Errorf("something went wrong in target %q : %q", id, err)
	}

	changelogSourceID := target.Config.SourceID
	if changelogSourceID == "" {
		switch len(sourceIds) {
		case 1:
			changelogSourceID = sourceIds[0]
		case 0:
		// If we have more than one sourceID then we can't define in a reliable way which one to use
		// as the order of the sourceIDs is not guaranteed.
		default:
			logrus.Debugf("Target depends on a too many sources that we can't determine which one to use for the changelog")
		}
	}

	if changelogSourceID != "" {
		// Once the source is executed, then it can retrieve its changelog
		// Any error means an empty changelog
		if source, found := p.Sources[changelogSourceID]; found {
			c, err := resource.New(source.Config.ResourceConfig)

			if err == nil {

				changelogs := c.Changelog(target.Result.Information, source.OriginalOutput)

				if changelogs != nil {
					target.Result.Changelogs = *changelogs

					logrus.Infof("%s", changelogs.String())

				} else {
					logrus.Debugln("no changelog detected")
				}
			}
		}
	}

	p.Targets[id] = target
	p.Report.Targets[id] = &target.Result

	return target.Result.Result, target.Result.Changed, err
}
