package engine

// Show displays configurations that should be apply.
func (e *Engine) Show() (err error) {

	for _, pipeline := range e.Pipelines {

		PrintTitle(pipeline.Config.Spec.Name)

		if e.Options.DisplayFlavour == "graph" {
			err = pipeline.Graph(e.Options.GraphFlavour)
		} else {
			err = pipeline.Config.Display()
		}
		if err != nil {
			return err
		}

	}
	return nil
}
