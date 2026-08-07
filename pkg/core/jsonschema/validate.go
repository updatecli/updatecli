package jsonschema

import (
	"errors"
	"fmt"
	"strings"

	validator "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

// Problem describes one document location that does not match a schema.
type Problem struct {
	// Path locates the problem within the document, such as "policies.0.polcy".
	Path string
	// Message explains the problem.
	Message string
}

func (p Problem) String() string {

	if p.Path == "" {
		return p.Message
	}

	return fmt.Sprintf("%s: %s", p.Path, p.Message)
}

// Validate checks a decoded document against a compiled schema and returns every problem
// found, each rooted at the location it was found in.
//
// It suits the documents described by a single schema. A document dispatching on a kind,
// such as an Updatecli manifest, is better validated one resource at a time so that a
// problem is not reported as the document matching none of the supported kinds.
func Validate(schema *validator.Schema, document interface{}) []Problem {

	err := schema.Validate(NormalizeValue(document))
	if err == nil {
		return nil
	}

	var validationErr *validator.ValidationError
	if !errors.As(err, &validationErr) {
		return []Problem{{Message: err.Error()}}
	}

	problems := []Problem{}
	flatten(validationErr, &problems)

	return problems
}

// flatten walks down to the leaf causes, which are the ones carrying both a precise
// location and a typed reason.
func flatten(err *validator.ValidationError, problems *[]Problem) {

	if len(err.Causes) > 0 {
		for _, cause := range err.Causes {
			flatten(cause, problems)
		}
		return
	}

	*problems = append(*problems, Problem{
		Path:    strings.Join(err.InstanceLocation, "."),
		Message: DescribeError(err),
	})
}

// DescribeError phrases a validation failure in terms of the document rather than of the
// schema keyword that rejected it.
func DescribeError(err *validator.ValidationError) string {

	switch errorKind := err.ErrorKind.(type) {

	case *kind.AdditionalProperties:
		messages := make([]string, len(errorKind.Properties))
		for i, property := range errorKind.Properties {
			messages[i] = fmt.Sprintf("unknown key %q", property)
		}
		return strings.Join(messages, ", ")

	case *kind.Required:
		return fmt.Sprintf("missing required key(s) %s", quoteAll(errorKind.Missing))

	case *kind.Type:
		return fmt.Sprintf("expected %s, got %s", strings.Join(errorKind.Want, " or "), errorKind.Got)

	case *kind.Enum:
		return fmt.Sprintf("value %v is not one of %v", errorKind.Got, errorKind.Want)

	default:
		return err.ErrorKind.LocalizedString(nil)
	}
}

func quoteAll(values []string) string {

	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = fmt.Sprintf("%q", value)
	}

	return strings.Join(quoted, ", ")
}

// NormalizeValue turns a decoded YAML tree into one a schema validator understands, which
// means mappings keyed by strings.
//
// A tree must be decoded with the same YAML library as the document itself rather than
// through a JSON round trip, which would turn every number into a float and report
// integer fields as invalid.
func NormalizeValue(value interface{}) interface{} {

	switch typed := value.(type) {

	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			normalized[key] = NormalizeValue(item)
		}
		return normalized

	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			normalized[fmt.Sprintf("%v", key)] = NormalizeValue(item)
		}
		return normalized

	case []interface{}:
		normalized := make([]interface{}, len(typed))
		for i, item := range typed {
			normalized[i] = NormalizeValue(item)
		}
		return normalized

	default:
		return value
	}
}
