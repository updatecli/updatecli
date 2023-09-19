package engine

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Show displays configurations that should be apply.
func (e *Engine) Show() error {

	for _, pipeline := range e.Pipelines {

		logrus.Infof("\n\n%s\n", strings.Repeat("#", len(pipeline.Config.Spec.Name)+4))
		logrus.Infof("# %s #\n", strings.ToTitle(pipeline.Config.Spec.Name))
		logrus.Infof("%s\n\n", strings.Repeat("#", len(pipeline.Config.Spec.Name)+4))

		err := pipeline.Config.Display()
		if err != nil {
			return err
		}

	}
	return nil
}
