package ami

import "github.com/olblak/updateCli/pkg/core/scm"

// Condition test if a image matching specific filters exist.
func (a *AMI) Condition(source string) (bool, error) {
	//
	return true, nil
}

// ConditionFromSCM is a placeholder to validate the condition interface
func (a *AMI) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	//
	return true, nil
}
