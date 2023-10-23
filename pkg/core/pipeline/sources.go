package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunSources iterates on every source definition to retrieve every information.
func (p *Pipeline) RunSources() error {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Sources"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Source")+1))

	sortedSourcesKeys, err := SortedSourcesKeys(&p.Sources)
	if err != nil {
		logrus.Errorf("%s %v\n", result.FAILURE, err)
		return err
	}

	for _, id := range sortedSourcesKeys {
		err = p.Update()
		if err != nil {
			return err
		}

		source := p.Sources[id]
		source.Config = p.Config.Spec.Sources[id]

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		shouldRunSource := true
		for _, parentSource := range source.Config.DependsOn {
			if p.Sources[parentSource].Result.Result != result.SUCCESS {
				logrus.Warningf("Parent source[%q] did not succeed. Skipping execution of the source[%q]", parentSource, id)
				shouldRunSource = false
			}
		}

		if !shouldRunSource {
			continue
		}

		err = source.Run()
		if err != nil {
			source.Result.Result = result.FAILURE

			p.Sources[id] = source
			p.Report.Sources[id] = &source.Result
			p.Report.Result = result.FAILURE

			logrus.Errorf("%s %v\n", source.Result, err)
			continue
		}

		if len(source.Changelog) > 0 {
			logrus.Infof("\n\n%s:\n", strings.ToTitle("Changelog"))
			logrus.Infof("%s\n", strings.Repeat("-", len("Changelog")+1))
			logrus.Infof("%s\n", source.Changelog)
		}

		p.Sources[id] = source
		p.Report.Sources[id] = &source.Result

		err = p.Report.UpdateResult(source.Result.Result)
		if err != nil {
			return fmt.Errorf("unable to set report result: %s", err)
		}
	}

	return err
}
