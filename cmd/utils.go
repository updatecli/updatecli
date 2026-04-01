package cmd

import (
	"strings"
)

// parseParametersList is a helper function to parse a list of parameters that can be passed as a comma separated list or as multiple values
func parseParametersList(input []string) []string {
	result := []string{}

	for i := range input {
		for j := range strings.Split(input[i], ",") {
			result = append(result, strings.TrimSpace(strings.Split(input[i], ",")[j]))
		}
	}

	return result
}
