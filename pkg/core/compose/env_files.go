package compose

import (
	"fmt"

	"github.com/joho/godotenv"
)

/*
	It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults.
*/

// EnvFiles is a list of environment files

type EnvFiles []string

// SetEnv sets the environment variables
func (e EnvFiles) SetEnv() error {
	if len(e) == 0 {
		return nil
	}
	err := godotenv.Load(e...)
	if err != nil {
		return fmt.Errorf("error setting environment variables: %s", err)
	}

	return nil
}
