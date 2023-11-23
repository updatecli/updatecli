package toml

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (t *Toml) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	conditionResult := true

	resultMessage := ""

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {

		if err := t.contents[i].Read(rootDir); err != nil {
			return false, "", fmt.Errorf("reading toml file: %w", err)
		}

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(t.spec.Query) > 0 {
		case true:
			queryResults, err = t.contents[i].MultipleQuery(t.spec.Query)
			if err != nil {
				return false, "", err
			}

		case false:
			queryResult, err := t.contents[i].Query(t.spec.Key)
			if err != nil {
				return false, "", err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == t.spec.Value {
			case true:
				resultMessage = fmt.Sprintf("%s\nkey %q, from file %q, is correctly set to %q",
					resultMessage,
					t.spec.Key,
					t.contents[i].FilePath,
					t.spec.Value)

			case false:
				conditionResult = false
				resultMessage = fmt.Sprintf("%s\nkey %q, from file %q, is incorrectly set to %q and should be %q",
					resultMessage,
					t.spec.Key,
					t.contents[i].FilePath,
					queryResult,
					t.spec.Value)
			}
		}

	}

	resultMessage = strings.TrimPrefix(resultMessage, "\n")

	return conditionResult, resultMessage, nil
}
