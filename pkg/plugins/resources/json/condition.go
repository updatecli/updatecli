package json

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (j *Json) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	conditionResult := true
	partialMessage := ""

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range j.contents {

		if err := j.contents[i].Read(rootDir); err != nil {
			return false, "", fmt.Errorf("reading json file: %w", err)
		}

		// Override value from source if not yet defined
		if len(j.spec.Value) == 0 {
			j.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(j.spec.Query) > 0 {
		case true:
			queryResults, err = j.contents[i].MultipleQuery(j.spec.Query)
			if err != nil {
				return false, "", err
			}

		case false:
			queryResult, err := j.contents[i].Query(j.spec.Key)
			if err != nil {
				return false, "", err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == j.spec.Value {
			case true:
				partialMessage = fmt.Sprintf("%s\nKey %q, from file %q, is correctly set to %q",
					partialMessage,
					j.spec.Key,
					j.contents[i].FilePath,
					j.spec.Value)

			case false:
				conditionResult = false
				partialMessage = fmt.Sprintf("%s\nKey %q, from file %q, is incorrectly set to %q and should be %q",
					partialMessage,
					j.spec.Key,
					j.contents[i].FilePath,
					queryResult,
					j.spec.Value)
			}
		}

	}

	partialMessage = strings.TrimPrefix(partialMessage, "\n")
	return conditionResult, partialMessage, nil
}
