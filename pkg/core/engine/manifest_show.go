package engine

import (
	"github.com/updatecli/updatecli/pkg/core/log"
)

// Show displays configurations that should be apply.
func (e *Engine) Show() (err error) {

	for _, pipeline := range e.Pipelines {

		graphOutput := ""

		log.Title(pipeline.Config.Spec.Name)

		if e.Options.DisplayFlavor == "graph" {
			graphOutput, err = pipeline.Graph(e.Options.GraphFlavor)
			log.Infof("%s", graphOutput)

		} else {
			err = pipeline.Config.Display()
		}
		if err != nil {
			return err
		}

	}
	return nil
}
