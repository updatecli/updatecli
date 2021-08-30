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
	if source == "" && len(a.Filters) == 0 {
		logrus.Infof("\u2717 No AMI could be found as no AMI filters defined\n")
		return false, nil
	}

	// Based on source information,
	// we try to define a default image-id resource
	// if not researched
	isImageIDDefined := false
	for i := 0; i < len(a.Filters); i++ {
		if strings.Compare(a.Filters[i].Name, "image-id") == 0 {
			isImageIDDefined = true
			break
		}

	}

	// Set image-id to source output if not yet defined
	if !isImageIDDefined {
		a.Filters = append(a.Filters, Filter{
			Name:   "image-id",
			Values: source,
		})
	}

	svc, errs := a.Init()

	if len(errs) > 0 {
		for _, err := range errs {
			logrus.Printf("%s\n", err.Error())
		}
		return false, errors.New("something went wrong while testing if the AWS AMI exist.")
	}

	result, err := a.getLatestAmiID(svc)

	if err != nil {
		return false, err
	}

	if len(result) > 0 {
		logrus.Infof("\u2714 AMI %q found\n", result)
		return true, nil
	}

	fmt.Printf("\u2717 No AMI found matching criteria in region %s\n", a.Region)

	return false, nil
}

// ConditionFromSCM is a placeholder to validate the condition interface
func (a *AMI) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	fmt.Printf("\u2717 Condition with SCM is not supported, please remove the scm block \n")

	return false, errors.New("condition with SCM is not supported")
}
