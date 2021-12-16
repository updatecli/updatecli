package awsami

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition tests if an image matching the specific filters exists.
func (a *AMI) Condition(source string) (bool, error) {

	// It's an error if the upstream source is empty and the user does not provide any filter
	// then it mean
	if source == "" && len(a.Spec.Filters) == 0 {
		logrus.Errorln(ErrNoFilter)
		return false, ErrSpecNotValid
	}

	isFilterDefined := func(filter string) (found bool) {
		for i := 0; i < len(a.Spec.Filters); i++ {
			if strings.Compare(a.Spec.Filters[i].Name, filter) == 0 {
				found = true
				break
			}
		}
		return found
	}

	// Set image-id to source output if not yet defined
	if !isFilterDefined("image-id") && len(source) > 0 {
		a.Spec.Filters = append(a.Spec.Filters, Filter{
			Name:   "image-id",
			Values: source,
		})
	}

	logrus.Debugf("Looking for latest AMI ID matching:\n  ---\n  %s\n  ---\n\n",
		strings.TrimRight(
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n  "))

	foundAMI, err := a.getLatestAmiID()

	if err != nil {
		return false, err
	}

	if len(foundAMI) > 0 {
		logrus.Infof("%s AMI %q found\n", result.SUCCESS, foundAMI)
		return true, nil
	}

	fmt.Printf("%s No AMI found matching criteria for region %s\n", result.FAILURE, a.Spec.Region)

	return false, nil
}

// ConditionFromSCM is a placeholder to validate the condition interface
func (a *AMI) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {

	fmt.Printf("%s Condition with SCM is not supported, please remove the scm block \n", result.FAILURE)

	return false, errors.New("condition with SCM is not supported")
}
