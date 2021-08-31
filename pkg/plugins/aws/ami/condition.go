package ami

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/scm"
)

// Condition tests if an image matching the specific filters exists.
func (a *AMI) Condition(source string) (bool, error) {

	// It's an error if the upstream source is empty and the user does not provide any filter
	// then it mean
	if source == "" && len(a.Spec.Filters) == 0 {
		logrus.Infof("\u2717 No AMI could be found as no AMI filters defined\n")
		return false, nil
	}

	// Based on source information,
	// we try to define a default image-id resource
	// if not researched
	isImageIDDefined := false
	for i := 0; i < len(a.Spec.Filters); i++ {
		if strings.Compare(a.Spec.Filters[i].Name, "image-id") == 0 {
			isImageIDDefined = true
			break
		}
	}

	// Set image-id to source output if not yet defined
	if !isImageIDDefined {
		a.Spec.Filters = append(a.Spec.Filters, Filter{
			Name:   "image-id",
			Values: source,
		})
	}

	svc, errs := a.Init()

	if len(errs) > 0 {
		return false, errors.New("something went wrong while testing if the AWS AMI exist")
	}

	logrus.Debugf("Looking for latest AMI ID matching:\n  ---\n  %s\n  ---\n\n",
		strings.TrimRight(
			strings.ReplaceAll(a.Spec.String(), "\n", "\n  "), "\n  "))

	result, err := a.getLatestAmiID(svc)

	if err != nil {
		return false, err
	}

	if len(result) > 0 {
		logrus.Infof("\u2714 AMI %q found\n", result)
		return true, nil
	}

	fmt.Printf("\u2717 No AMI found matching criteria in region %s\n", a.Spec.Region)

	return false, nil
}

// ConditionFromSCM is a placeholder to validate the condition interface
func (a *AMI) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	fmt.Printf("\u2717 Condition with SCM is not supported, please remove the scm block \n")

	return false, errors.New("condition with SCM is not supported")
}
