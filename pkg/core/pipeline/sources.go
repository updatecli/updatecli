package pipeline

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func (p *Pipeline) updateSource(id, result string) {
	source := p.Sources[id]
	source.Result.Result = result
	p.Sources[id] = source
	p.Report.Sources[id] = &source.Result
}

func (p *Pipeline) RunSource(id string) (r string, err error) {
	source := p.Sources[id]
	source.Config = p.Config.Spec.Sources[id]
	source.Result.Name = source.Config.ResourceConfig.Name
	err = source.Run()
	if len(source.Changelog) > 0 {
		logrus.Infof("\n\n%s:\n", strings.ToTitle("Changelog"))
		logrus.Infof("%s\n", strings.Repeat("-", len("Changelog")+1))
		logrus.Infof("%s\n", source.Changelog)
	}
	p.Sources[id] = source
	return source.Result.Result, err
}
