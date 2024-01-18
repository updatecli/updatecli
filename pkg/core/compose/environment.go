package compose

import (
	"fmt"
	"os"
)

// Environments is a map of environment variables
type Environments map[string]string

// SetEnv sets the environment variables
func (e Environments) SetEnv() error {
	var errs []error
	for key, value := range e {
		if err := os.Setenv(key, value); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error setting environment variables: %s", errs)
	}

	return nil
}
