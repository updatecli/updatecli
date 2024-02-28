package pipeline

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// RunConditions run every conditions for a given configuration config.
func (p *Pipeline) RunConditions() (err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&p.Conditions)
	if err != nil {
		return err
	}

	for _, id := range sortedConditionsKeys {
		// Update pipeline before each condition run
		err = p.Update()
		if err != nil {
			return err
		}

		condition := p.Conditions[id]
		condition.Config = p.Config.Spec.Conditions[id]

		// Ensure the result named contains the up to date condition name after templating
		condition.Result.Name = condition.Config.ResourceConfig.Name

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		err := condition.Run(p.Sources[condition.Config.SourceID].Output)
		if err != nil {
			// Show error to end user if any but continue the flow execution
			logrus.Error(err)
		}

		p.Conditions[id] = condition
		p.Report.Conditions[id] = &condition.Result

	}

	return nil
}
