package version

import (
	"errors"
	"fmt"
)

var (
	// ErrNoVersionFound return a error when no version couldn't be found
	ErrNoVersionFound error = errors.New("no version found")
	// ErrNoVersionsFound return a error when the versions list is empty
	ErrNoVersionsFound error = errors.New("versions list empty")
	// ErrNoValidSemVerFound return a error when the versions list is empty
	ErrNoValidSemVerFound error = errors.New("no valid semantic version found")
	// ErrNoValidDateFound return a error when the versions list is empty
	ErrNoValidDateFound error = errors.New("no valid date found")
)

// ErrNoVersionFoundForPattern returns when a given pattern does not find any version
type ErrNoVersionFoundForPattern struct {
	Pattern string
}

func (e *ErrNoVersionFoundForPattern) Error() string {
	return fmt.Sprintf("no version found matching pattern %q", e.Pattern)
}

// ErrUnsupportedVersionKind returns when the provided version filter is unsupported
type ErrUnsupportedVersionKind struct {
	Kind string
}

func (e *ErrUnsupportedVersionKind) Error() string {
	return fmt.Sprintf("unsupported version kind %q", e.Kind)
}

// ErrUnsupportedVersionKind returns when the provided combination of version and pattern filter is unsupported
type ErrUnsupportedVersionKindPattern struct {
	Kind    string
	Pattern string
}

func (e *ErrUnsupportedVersionKindPattern) Error() string {
	return fmt.Sprintf("unsupported version kind %q with pattern %q", e.Kind, e.Pattern)
}

// ErrUnsupportedVersionKind returns when the provided combination of version and pattern filter is unsupported
type ErrIncorrectSemVerConstraint struct {
	SemVerConstraint string
}

func (e *ErrIncorrectSemVerConstraint) Error() string {
	return fmt.Sprintf("wrong semantic versioning constraint %q", e.SemVerConstraint)
}
