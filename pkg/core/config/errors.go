package config

import "errors"

var (
	// ErrConfigFileTypeNotSupported is returned when updatecli try to read
	// an unsupported file type.
	ErrConfigFileTypeNotSupported = errors.New("file extension not supported")

	// ErrBadConfig is returned when updatecli try to read
	// a wrong configuration.
	ErrBadConfig = errors.New("wrong updatecli configuration")

	// ErrNoEnvironmentVariableSet is returned when during the templating process,
	// it tries to access en environment variable not set.
	ErrNoEnvironmentVariableSet = errors.New("environment variable doesn't exist")

	// ErrNoKeyDefined is returned when during the templating process, it tries to
	// retrieve a key value which is not defined in the configuration
	ErrNoKeyDefined = errors.New("key not defined in configuration")

	// ErrNotAllowedTemplatedKey is returned when
	// we are planning to template at runtime unauthorized keys such map key
	ErrNotAllowedTemplatedKey = errors.New("not allowed templated key")
)
