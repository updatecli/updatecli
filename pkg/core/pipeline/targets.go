package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrRunTargets is return when at least one error happened during targets execution
	ErrRunTargets error = errors.New("something went wrong during target execution")
)

// RunTargets iterates on every target to update each of them.
func (p *Pipeline) RunTargets() error {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Targets"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Targets")+1))

	// Sort targets keys by building a dependency graph
	sortedTargetsKeys, err := SortedTargetsKeys(&p.Targets)
	if err != nil {
		p.Report.Result = result.FAILURE
		return fmt.Errorf("sort targets: %w", err)
	}

	isResultChanged := false
	p.Report.Result = result.SUCCESS

	errs := []error{}

	for _, id := range sortedTargetsKeys {
		// Update pipeline before each target run
		err = p.Update()
		if err != nil {
			return fmt.Errorf("update pipeline: %w", err)
		}

		logrus.Infof("\n%s\n", id)
		logrus.Infof("%s\n", strings.Repeat("-", len(id)))

		target := p.Targets[id]
		target.Config = p.Config.Spec.Targets[id]

		// Ensure the result named contains the up to date target name after templating
		target.Result.Name = target.Config.ResourceConfig.Name
		target.Result.DryRun = target.DryRun

		shouldSkipTarget := false

		dependsOnMatchingTarget := p.isDependsOnMatchingTarget(&target)
		if dependsOnMatchingTarget != "" {
			logrus.Debugf("Skipping target[%q] because of dependsOn conditions", id)
			target.Result.ConsoleOutput = target.Result.ConsoleOutput + dependsOnMatchingTarget
			shouldSkipTarget = true
			target.Result.Result = result.SKIPPED
		}

		// Check if the target should be skipped
		if !target.Config.DisableConditions && !shouldSkipTarget {
			failingConditions := []string{}

			switch len(target.Config.ConditionIDs) > 0 {
			case true:
				for _, conditionID := range target.Config.ConditionIDs {
					if p.Conditions[conditionID].Result.Result != result.SUCCESS {
						failingConditions = append(failingConditions, conditionID)
					}
				}
			case false:
				// if no condition is defined, we evaluate all conditions
				for id, condition := range p.Conditions {
					if condition.Result.Result != result.SUCCESS {
						failingConditions = append(failingConditions, id)
					}
				}
			}

			if len(failingConditions) > 0 {
				target.Result.ConsoleOutput = fmt.Sprintf("Conditions %v not met. Skipping execution of the target[%q]", failingConditions, id)
				logrus.Infof(target.Result.ConsoleOutput)
				shouldSkipTarget = true
			}
		}

		// No need to run this target as one of its dependency failed
		if shouldSkipTarget {
			p.Targets[id] = target
			p.Report.Targets[id] = &target.Result
			continue
		}

		err = target.Run(p.Sources[target.Config.SourceID].Output, &p.Options.Target)

		if err != nil {
			p.Report.Result = result.FAILURE
			target.Result.Result = result.FAILURE

			errs = append(errs, fmt.Errorf("something went wrong in target %q : %q", id, err))
		}

		p.Targets[id] = target
		p.Report.Targets[id] = &target.Result

		if target.Result.Changed {
			isResultChanged = target.Result.Changed
		}
	}

	if len(errs) > 0 {
		logrus.Infof("\n")
		for _, e := range errs {
			logrus.Errorln(e)
		}
		logrus.Infof("\n")

		p.Report.Result = result.FAILURE
		return ErrRunTargets
	}

	if isResultChanged {
		p.Report.Result = result.ATTENTION
	}

	return nil
}

// isDependsOnMatchingTarget checks if the target dependsOn conditions are met.
// if not, the target is skipped.
// a dependsOn value must follow one of the three following format:
// dependsOn:
//   - targetID
//   - targetID:and
//   - targetID:or
//
// where targetID is the id of another target in the manifest
// "and" the boolean operator is optional and can be used to specify that all conditions must be met
// "or" the boolean operator is optional and can be used to specify that at least one condition must be met
// if the boolean operator is not provided, it defaults to "and"
func (p *Pipeline) isDependsOnMatchingTarget(target *target.Target) string {
	// exit early
	if len(target.Config.DependsOn) == 0 {
		return ""
	}

	failingRequiredDependsOn := []string{}
	failingOptionalDependsOn := []string{}
	counterRequiredDependsOn := 0
	counterOptionalDependsOn := 0

	andOperator := "and"
	orOperator := "or"

	for _, parentTarget := range target.Config.DependsOn {
		var booleanOperator string

		parentTarget, booleanOperator = parseDependsOnValue(parentTarget)

		//parentTargetArray := strings.Split(parentTarget, ":")
		//switch len(parentTargetArray) {
		//case 1:
		//	parentTarget = parentTargetArray[0]
		//	booleanOperator = andOperator
		//case 2:
		//	parentTarget = parentTargetArray[0]
		//	booleanOperator = parentTargetArray[1]
		//default:
		//	counterRequiredDependsOn++
		//	failingRequiredDependsOn = append(
		//		failingRequiredDependsOn,
		//		fmt.Sprintf("Invalid dependsOn format %q", parentTarget),
		//	)
		//	continue
		//}

		switch booleanOperator {
		case andOperator:
			counterRequiredDependsOn++
		case orOperator:
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
		if p.Targets[parentTarget].Result.Result == result.FAILURE {
			switch booleanOperator {
			case andOperator:
				failingRequiredDependsOn = append(
					failingRequiredDependsOn,
					fmt.Sprintf("Required Parent target[%q] did not succeed.", parentTarget))
			case orOperator:
				failingOptionalDependsOn = append(
					failingRequiredDependsOn,
					fmt.Sprintf("Parent target[%q] did not succeed.", parentTarget))
			}
		} else if target.Config.DependsOnChange && p.Targets[parentTarget].Result.Changed {
			switch booleanOperator {
			case andOperator:
				failingRequiredDependsOn = append(failingRequiredDependsOn,
					fmt.Sprintf("Required Parent target[%q] changed.", parentTarget),
				)
			case orOperator:
				failingOptionalDependsOn = append(failingRequiredDependsOn,
					fmt.Sprintf("Parent target[%q] changed.", parentTarget),
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
