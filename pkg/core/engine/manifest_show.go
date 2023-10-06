package engine

// Show displays configurations that should be apply.
func (e *Engine) Show() error {

	for _, pipeline := range e.Pipelines {

		PrintTitle(pipeline.Config.Spec.Name)

		err := pipeline.Config.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
