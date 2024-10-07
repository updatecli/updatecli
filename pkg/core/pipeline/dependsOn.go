package pipeline

import (
	"strings"
)

// parseDependsOnValue split the input string into category, key and booleanOperator
func parseDependsOnValue(val string) (key, booleanOperator, category string) {
	if val == "" {
		return "", "", ""
	}

	catArray := strings.Split(val, "#")
	switch len(catArray) {
	case 1:
		// Nothing to do
	default:
		category = catArray[0]
		val = strings.Join(catArray[1:], "#")

	}
	valArray := strings.Split(val, ":")

	switch len(valArray) {
	case 1:
		return valArray[0], andBooleanOperator, category
	default:
		return strings.Join(valArray[:len(valArray)-1], ":"), valArray[len(valArray)-1], category
	}
}

// isValidateDependsOn test if we are referencing an exist resource key
func isValidDependsOn(dependsOn string, index map[string]string) bool {

	for val := range index {
		// some dependsOn accept a value after a colon
		// which affect the behavior
		// such as "targetName:or"
		dependsOn, _, _ = parseDependsOnValue(dependsOn)

		switch dependsOn {
		case val:
			return true
		case "":
			return false
		}
	}
	return false
}
