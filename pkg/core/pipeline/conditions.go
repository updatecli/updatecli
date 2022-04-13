package pipeline

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// RunConditions run every conditions for a given configuration config.
func (p *Pipeline) RunConditions() (globalResult bool, err error) {

	if len(p.Conditions) == 0 {
		logrus.Debugln("No conditions to run")
		return true, nil
	}

	logrus.Infof("\n\n%s:\n", strings.ToTitle("conditions"))
	logrus.Infof("%s\n", strings.Repeat("=", len("conditions")+1))

	// Sort conditions keys by building a dependency graph
	sortedConditionsKeys, err := SortedConditionsKeys(&p.Conditions)
	if err != nil {
		return false, err
	}

	globalResult = true

	for _, id := range sortedConditionsKeys {
		// Update pipeline before each condition run
		err = p.Config.Update(p)
		if err != nil {
			globalResult = false
			return globalResult, err
		}

		condition := p.Conditions[id]
		condition.Config = p.Config.Conditions[id]

		rpt := p.Report.Conditions[id]
		// Update report name as the condition configuration might has been updated (templated values)
		rpt.Name = condition.Config.Name

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		err := condition.Run(p.Sources[condition.Config.SourceID].Output)
		if err != nil {
			// Show error to end user if any but continue the flow execution
			logrus.Error(err)
		}

		// Reports the result of the execution of this condition
		rpt.Result = condition.Result

		// If there was an error OR if the condition is not successful then defines the global result as false
		if err != nil || condition.Result != result.SUCCESS {
			globalResult = false
		}

		p.Conditions[id] = condition
		p.Report.Conditions[id] = rpt

	}

	return globalResult, nil
}
