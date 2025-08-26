package engine

import (
	"github.com/sirupsen/logrus"
)

// Show displays configurations that should be apply.
func (e *Engine) Show() (err error) {

	for _, pipeline := range e.Pipelines {

		graphOutput := ""

		PrintTitle(pipeline.Config.Spec.Name)

		if e.Options.DisplayFlavor == "graph" {
			graphOutput, err = pipeline.Graph(e.Options.GraphFlavor)
			logrus.Infof("%s", graphOutput)

		} else {
			err = pipeline.Config.Display()
		}
		if err != nil {
			return err
		}

	}
	return nil
}
