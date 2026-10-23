package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// updatecliFuncMap returns a map of functions used by updatecli at init time, it will ignore func used
// at runtime
func updatecliFuncMap() template.FuncMap {

	return template.FuncMap{
		// Retrieve value from environment variable, return error if not found
		"requiredEnv": func(s string) (string, error) {
			value := os.Getenv(s)
			if value == "" {
				return "", errors.New("no value found for environment variable " + s)
			}
			return value, nil
		},
		"pipeline": func(s string) (string, error) {
			return fmt.Sprintf(`{{ pipeline %q }}`, s), nil
		},
		"source": func(s string) (string, error) {
			return fmt.Sprintf(`{{ source %q }}`, s), nil
		},
	}
}

// updatecliRuntimeFuncMap returns a map of functions used by updatecli at runtime time.
func updatecliRuntimeFuncMap(data interface{}) template.FuncMap {
	return template.FuncMap{
		"pipeline": func(s string) (string, error) {
			/*
				Retrieve the value of a third location key from
				the updatecli configuration.
				It returns an error if a key doesn't exist
				It returns {{ pipeline "<key>" }} if a key exist but still set to zero value,
				then we assume that the value will be set later in the run.
				Otherwise it returns the value.
				This func is design to constantly reevaluate if a configuration changed
			*/

			val, err := getFieldValueByQuery(data, strings.Split(s, "."))
			if err != nil {
				return "", err
			}

			if len(val) > 0 {
				return val, nil
			}
			// If we couldn't find a value, then we return the function so we can retry
			// later on.
			return fmt.Sprintf("{{ pipeline %q }}", s), nil

		},
		"source": func(s string) (string, error) {
			/*
				Retrieve the value of a third location key from
				the updatecli context.
				It returns an error if a key doesn't exist
				It returns {{ source "<key>" }} if a key exist but still set to zero value,
				then we assume that the value will be set later in the run.
				Otherwise it returns the value.
				This func is design to constantly reevaluate if a configuration changed
			*/

			sourceResult, err := getFieldValueByQuery(data, []string{"Sources", s, "Result", "Result"})
			if err != nil {
				return "", err
			}

			switch sourceResult {
			case result.SUCCESS:
				return getFieldValueByQuery(data, []string{"Sources", s, "Output"})
			case result.FAILURE:
				return "", fmt.Errorf("parent source %q failed", s)
			case result.SKIPPED:
				// A source is initialized with a SKIPPED result, so SKIPPED alone
				// is ambiguous: it can mean "not run yet" or "explicitly skipped".
				// IsRun disambiguates the two. A source still pending execution
				// (IsRun false) is returned literally so it can be re-evaluated on
				// a later configuration iteration, while a source that the pipeline
				// finalized as skipped (IsRun true) is a hard error.
				isRun, err := getFieldBoolByQuery(data, []string{"Sources", s, "Result", "IsRun"})
				if err != nil {
					return "", err
				}
				if isRun {
					return "", fmt.Errorf("parent source %q was skipped", s)
				}
				return fmt.Sprintf("{{ source %q }}", s), nil
			// If the result of the parent source execution is not SUCCESS, FAILURE, or SKIPPED,
			// then it means it was not already run.
			// In this case, the function is returned "as it" (literally) to allow retry later (on a second configuration iteration)
			default:
				return fmt.Sprintf("{{ source %q }}", s), nil
			}
		},
	}
}
