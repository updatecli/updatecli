package pipeline

func (p *Pipeline) updateSource(id, result string) {
	source := p.Sources[id]
	source.Result.Result = result
	p.Sources[id] = source
	p.Report.Sources[id] = &source.Result
}

func (p *Pipeline) RunSource(id string) (r string, err error) {
	source := p.Sources[id]
	source.Config = p.Config.Spec.Sources[id]
	source.Result.Name = source.Config.Name

	err = source.Run()

	p.Sources[id] = source

	return source.Result.Result, err
}
