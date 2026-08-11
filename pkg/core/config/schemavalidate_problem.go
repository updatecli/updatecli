package config

import (
	"fmt"
	"strings"
)

// Severity qualifies how much a schema problem matters.
type Severity string

const (
	// SeverityError reports something Updatecli cannot interpret.
	SeverityError Severity = "error"
	// SeverityWarning reports something Updatecli still accepts, such as a deprecated
	// key, or something it cannot fully check.
	SeverityWarning Severity = "warning"
)

// SchemaProblem describes one manifest location that does not match the Updatecli schema.
type SchemaProblem struct {
	// File is the manifest path as provided by the user.
	File string
	// Document is the index of the YAML document within File, manifests may hold several.
	Document int
	// Path locates the problem within the document, such as
	// "targets.mytarget.spec.versionfilter.kind".
	Path string
	// Severity qualifies the problem.
	Severity Severity
	// Message explains the problem.
	Message string
}

func (p SchemaProblem) String() string {

	location := p.File
	if p.Document > 0 {
		location = fmt.Sprintf("%s[%d]", location, p.Document)
	}

	if p.Path != "" {
		location = fmt.Sprintf("%s %s", location, p.Path)
	}

	return fmt.Sprintf("%s: %s", location, p.Message)
}

// SchemaReport gathers every problem found in a manifest.
type SchemaReport struct {
	File     string
	Problems []SchemaProblem
}

// HasErrors reports whether at least one problem is an error.
func (r SchemaReport) HasErrors() bool {

	for _, problem := range r.Problems {
		if problem.Severity == SeverityError {
			return true
		}
	}

	return false
}

// Errors returns the problems of error severity.
func (r SchemaReport) Errors() []SchemaProblem {
	return r.filter(SeverityError)
}

// Warnings returns the problems of warning severity.
func (r SchemaReport) Warnings() []SchemaProblem {
	return r.filter(SeverityWarning)
}

func (r SchemaReport) filter(severity Severity) []SchemaProblem {

	problems := []SchemaProblem{}
	for _, problem := range r.Problems {
		if problem.Severity == severity {
			problems = append(problems, problem)
		}
	}

	return problems
}

// Error implements the error interface so that a report can be returned where the
// manifest loading code already expects a multi line error.
func (r SchemaReport) Error() string {

	messages := []string{}
	for _, problem := range r.Errors() {
		messages = append(messages, problem.String())
	}

	return strings.Join(messages, "\n")
}
