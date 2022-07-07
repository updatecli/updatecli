package shell

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Environments []Environment

func (e Environments) ToStringSlice() []string {
	result := make([]string, len(e))

	for i, environment := range e {
		result[i] = environment.String()
	}

	return result

}

func (e Environments) isDuplicate() bool {
	foundName := map[string]struct{}{}
	foundDuplicatedName := []string{}

	for _, env := range e {
		if _, ok := foundName[env.Name]; ok {
			foundDuplicatedName = append(foundDuplicatedName, env.Name)
			continue
		}

		foundName[env.Name] = struct{}{}
	}

	if len(foundDuplicatedName) > 0 {
		logrus.Warningf("duplicated environment variable found: [%q]\n", strings.Join(foundDuplicatedName, ","))
		return true
	}

	return false
}

// Ensures that we don't have duplicated value for a variable and that the user is not attempting to override the DRY_RUN reserved variable.
func (e Environments) Validate() error {

	gotErr := false
	for _, environment := range e {
		err := environment.Update()
		if err != nil {
			logrus.Errorf("validating environment variable %q", environment.Name)
			gotErr = true
		}

		if environment.Name == DryRunVariableName {
			gotErr = true
			logrus.Errorf("environment variable %q is defined and overidden by the Updatecli process", DryRunVariableName)

		}
	}

	if duplicate := e.isDuplicate(); duplicate {
		gotErr = true
	}

	if gotErr {
		return fmt.Errorf("wrong configuration")
	}
	return nil
}
