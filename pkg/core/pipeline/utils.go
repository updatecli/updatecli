package pipeline

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// isDependsOnMatchingResource checks if the resource dependsOn conditions are met.
// if not, the resource is skipped.
// a dependsOn value must follow one of the three following format:
// dependsOn:
//   - (resourceType#)resourceID
//   - (resourceType#)resourceID:and
//   - (resourceType#)resourceID:or
//
// where `resourceType` is the type of resource of the dependent, if none are specified, it defaults to its own type
// `resourceID` is the id of another resource in the manifest
// "and" the boolean operator is optional and can be used to specify that all conditions must be met
// "or" the boolean operator is optional and can be used to specify that at least one condition must be met
// if the boolean operator is not provided, it defaults to "and"
func (p *Pipeline) isDependsOnMatchingResource(leaf *Node, depsResults map[string]*Node) string {
	// exit early
	if len(leaf.DependsOn) == 0 {
		return ""
	}

	failingRequiredDependsOn := []string{}
	failingOptionalDependsOn := []string{}
	counterRequiredDependsOn := 0
	counterOptionalDependsOn := 0

	for _, dependency := range leaf.DependsOn {
		dependencyResult := depsResults[dependency.ID]
		booleanOperator := dependency.Operator

		switch booleanOperator {
		case andBooleanOperator:
			counterRequiredDependsOn++
		case orBooleanOperator:
			counterOptionalDependsOn++
		default:
			counterRequiredDependsOn++
			failingRequiredDependsOn = append(
				failingRequiredDependsOn,
				fmt.Sprintf("Invalid boolean operator %q", booleanOperator),
			)
			continue
		}

		// We always fail first if the dependsOn target failed.
		if dependencyResult.Result == result.FAILURE {
			switch booleanOperator {
			case andBooleanOperator:
				failingRequiredDependsOn = append(
					failingRequiredDependsOn,
					fmt.Sprintf("Required Parent %q did not succeed.", dependency))
			case orBooleanOperator:
				failingOptionalDependsOn = append(
					failingRequiredDependsOn,
					fmt.Sprintf("Parent %q did not succeed.", dependency))
			}
			//}
		} else if leaf.DependsOnChange && dependencyResult.Changed {
			switch booleanOperator {
			case andBooleanOperator:
				failingRequiredDependsOn = append(failingRequiredDependsOn,
					fmt.Sprintf("Required Parent %q changed.", dependency),
				)
			case orBooleanOperator:
				failingOptionalDependsOn = append(failingRequiredDependsOn,
					fmt.Sprintf("Parent %q changed.", dependency),
				)
			}
		}
	}

	if counterRequiredDependsOn > 0 {
		if counterRequiredDependsOn == len(failingRequiredDependsOn) {
			logrus.Debugf(
				"All required dependsOn matched:\n\t* %s",
				strings.Join(failingRequiredDependsOn, "\n\t *"))
			return ""
		}
		logrus.Debugf(
			"%v required dependsOn matched over %v:\n\t* %s",
			len(failingRequiredDependsOn),
			counterRequiredDependsOn,
			strings.Join(failingRequiredDependsOn, "\n\t* "),
		)
	}

	if len(failingOptionalDependsOn) > 0 {
		logrus.Debugf(
			"%v optional dependsOn matched over %v:\n\t* %s",
			len(failingOptionalDependsOn),
			counterOptionalDependsOn,
			strings.Join(failingOptionalDependsOn, "\n\t* "),
		)
		return ""
	}

	return "Dependson not met."
}
