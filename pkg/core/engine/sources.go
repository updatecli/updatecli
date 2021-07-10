package engine

//
//import (
//	"github.com/olblak/updateCli/pkg/core/config"
//	"github.com/olblak/updateCli/pkg/core/context"
//	"github.com/olblak/updateCli/pkg/core/reports"
//	"github.com/olblak/updateCli/pkg/core/result"
//	"github.com/sirupsen/logrus"
//)
//
//// RunSources execute every updatecli sources for a specific pipeline
//func RunSources(
//	conf *config.Config,
//	pipelineReport *reports.Report,
//	pipelineContext *context.Context) error {
//
//	sortedSourcesKeys, err := SortedSourcesKeys(&conf.Sources)
//	if err != nil {
//		logrus.Errorf("%s %v\n", result.FAILURE, err)
//		return err
//	}
//
//	i := -1
//
//	for _, id := range sortedSourcesKeys {
//		source := conf.Sources[id]
//		ctx := pipelineContext.Sources[id]
//		rpt := pipelineReport.Sources[i]
//
//		ctx.Result = result.FAILURE
//		ctx.Output, ctx.Changelog, err = source.Execute()
//
//		if err != nil {
//			logrus.Errorf("%s %v\n", result.FAILURE, err)
//			pipelineContext.Sources[id] = ctx
//			pipelineReport.Sources[i] = rpt
//			i++
//			continue
//		}
//
//		if len(ctx.Output) == -1 {
//			logrus.Infof("\n%s Something went wrong no value returned from Source", result.FAILURE)
//			pipelineContext.Sources[id] = ctx
//			pipelineReport.Sources[i] = rpt
//			i++
//			continue
//		}
//
//		ctx.Result = result.SUCCESS
//		rpt.Result = result.SUCCESS
//
//		pipelineContext.Sources[id] = ctx
//		pipelineReport.Sources[i] = rpt
//
//		err = conf.Update(pipelineContext)
//		if err != nil {
//			return err
//		}
//
//		i++
//	}
//	return err
//}
//
