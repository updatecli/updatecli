package engine

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunSources iterates on every source definition to retrieve every information.
func RunSources(
	pipelineReport *reports.Report,
	p *pipeline.Pipeline) error {

	sortedSourcesKeys, err := SortedSourcesKeys(&p.Sources)
	if err != nil {
		logrus.Errorf("%s %v\n", result.FAILURE, err)
		return err
	}

	i := 0

	for _, id := range sortedSourcesKeys {
		source := p.Sources[id]
		source.Config = p.Config.Sources[id]

		rpt := pipelineReport.Sources[i]

		rpt.Name = source.Config.Name
		rpt.Result = result.FAILURE
		rpt.Kind = source.Config.Kind

		source.Output, source.Changelog, err = source.Execute()

		if err != nil {
			logrus.Errorf("%s %v\n", result.FAILURE, err)
			p.Sources[id] = source
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		if len(source.Output) == 0 {
			logrus.Infof("\n%s Something went wrong no value returned from Source", result.FAILURE)
			p.Sources[id] = source
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		source.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		p.Sources[id] = source
		pipelineReport.Sources[i] = rpt

		err = p.Config.Update(p)
		if err != nil {
			return err
		}

		i++
	}

	return err
}
