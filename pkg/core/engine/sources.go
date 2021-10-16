package engine

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/context"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunSources iterates on every source definition to retrieve every information.
func RunSources(
	pipelineReport *reports.Report,
	pipelineContext *context.Context) error {

	sortedSourcesKeys, err := SortedSourcesKeys(&pipelineContext.Sources)
	if err != nil {
		logrus.Errorf("%s %v\n", result.FAILURE, err)
		return err
	}

	i := 0

	for _, id := range sortedSourcesKeys {
		source := pipelineContext.Sources[id]
		source.Spec = pipelineContext.Config.Sources[id].Spec

		rpt := pipelineReport.Sources[i]

		rpt.Name = source.Spec.Name
		rpt.Result = result.FAILURE
		rpt.Kind = source.Spec.Kind

		source.Output, source.Changelog, err = source.Execute()

		if err != nil {
			logrus.Errorf("%s %v\n", result.FAILURE, err)
			pipelineContext.Sources[id] = source
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		if len(source.Output) == 0 {
			logrus.Infof("\n%s Something went wrong no value returned from Source", result.FAILURE)
			pipelineContext.Sources[id] = source
			pipelineReport.Sources[i] = rpt
			i++
			continue
		}

		source.Result = result.SUCCESS
		rpt.Result = result.SUCCESS

		pipelineContext.Sources[id] = source
		pipelineReport.Sources[i] = rpt

		err = pipelineContext.Config.Update(pipelineContext)
		if err != nil {
			return err
		}

		i++
	}

	return err
}
