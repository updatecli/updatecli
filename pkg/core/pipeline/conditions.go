package pipeline

func (p *Pipeline) updateCondition(id, result string) {
	condition := p.Conditions[id]
	condition.Result.Result = result
	p.Conditions[id] = condition
}

func (p *Pipeline) RunCondition(id string) (r string, err error) {
	condition := p.Conditions[id]
	condition.Config = p.Config.Spec.Conditions[id]
	condition.Result.Name = condition.Config.Name
	err = condition.Run(p.Sources[condition.Config.SourceID].Output)

	p.Conditions[id] = condition

	return condition.Result.Result, err
}
