package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const DisableChangelogEnvVar = "UPDATECLI_DISABLE_CHANGELOG"

const DisableVersionCheckEnvVar = "UPDATECLI_DISABLE_VERSION_CHECK"

// getEnvBoolOrDefault reads a boolean environment variable.
// It returns defaultValue when the variable is unset or invalid.
func getEnvBoolOrDefault(envVar string, defaultValue bool) bool {
	value, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		logrus.Debugf(
			"invalid boolean value for environment variable %q: %q, defaulting to %t",
			envVar,
			value,
			defaultValue,
		)
		return defaultValue
	}

	return parsed
}
