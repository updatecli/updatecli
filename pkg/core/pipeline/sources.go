package pipeline

import "context"

func (p *Pipeline) updateSource(id, result string) {
	source := p.Sources[id]
	source.Result.Result = result
	p.Sources[id] = source
}

func (p *Pipeline) RunSource(ctx context.Context, id string) (r string, err error) {
	source := p.Sources[id]
	source.Config = p.Config.Spec.Sources[id]
	source.Result.Name = source.Config.Name

	err = source.Run(ctx, p.SourceCache)

	p.Sources[id] = source

	return source.Result.Result, err
}
