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
func (a *AMI) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("condition with SCM is not supported, please remove the scm block")
		return errors.New("condition with SCM is not supported")
	}

	// It's an error if the upstream source is empty and the user does not provide any filter
	// then it mean
	if source == "" && len(a.Spec.Filters) == 0 {
		return ErrNoFilter
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
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n "))

	foundAMI, err := a.getLatestAmiID()

	if err != nil {
		return fmt.Errorf("getting latest AMI ID: %w", err)
	}

	if len(foundAMI) > 0 {
		resultCondition.Description = fmt.Sprintf("AMI %q found\n", foundAMI)
		resultCondition.Result = result.SUCCESS
		resultCondition.Pass = true
		return nil
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no AMI found matching criteria for region %s\n", a.Spec.Region)

	return nil
}
