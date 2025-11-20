package transformer

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	daselV2 "github.com/tomwright/dasel/v2"
)

var (
	// ErrEmptyInput is returned when we try to modify an empty value
	ErrEmptyInput = errors.New("validation error: transformer input is empty")
)

type JsonMatch struct {
	Key string `yaml:",omitempty" jsonschema:"required"`
	// If we don't find a match then return the following string or the input value
	NoMatchResult string `yaml:",omitempty"`
	// If we find multiple matches, join them by this
	JoinMultipleMatches string `yaml:",omitempty"`
	// If we find multiple matches, select the "first" or the "last"
	MultipleMatchSelector string `yaml:",omitempty"`
}

// Transformer holds a transformer rule
type Transformer struct {
	// AddPrefix adds a prefix to the transformer input value
	AddPrefix           string `yaml:",omitempty"`
	DeprecatedAddPrefix string `yaml:"addPrefix,omitempty" jsonschema:"-"`
	// AddSuffix adds a suffix to the transformer input value
	AddSuffix           string `yaml:",omitempty"`
	DeprecatedAddSuffix string `yaml:"addSuffix,omitempty" jsonschema:"-"`
	// TrimPrefix removes a prefix to the transformer input value
	TrimPrefix           string `yaml:",omitempty"`
	DeprecatedTrimPrefix string `yaml:"trimPrefix,omitempty" jsonschema:"-"`
	// TrimSuffix removes the suffix from the transformer input value
	TrimSuffix           string `yaml:",omitempty"`
	DeprecatedTrimSuffix string `yaml:"trimSuffix,omitempty" jsonschema:"-"`
	// Replacers specifies a list of replacer instruction
	Replacers Replacers `yaml:",omitempty"`
	// Replacer specifies what value needs to be changed and how
	Replacer Replacer `yaml:",omitempty"`
	// Find searches for a specific value if it exists and return false if it doesn't
	Find string `yaml:",omitempty"`
	// Find searches for a specific value if it exists then return the value using regular expression
	FindSubMatch           FindSubMatch `yaml:",omitempty"`
	DeprecatedFindSubMatch interface{}  `yaml:"findSubMatch,omitempty" jsonschema:"-"`
	JsonMatch              JsonMatch    `yaml:",omitempty"`
	// SemvVerInc specifies a comma separated list semantic versioning component that needs to be upgraded.
	SemVerInc           string `yaml:",omitempty"`
	DeprecatedSemVerInc string `yaml:"semverInc,omitempty" jsonschema:"-"`
	// Quote add quote around the value
	Quote bool `yaml:",omitempty"`
	// Unquote remove quotes around the value
	Unquote bool `yaml:",omitempty"`
}

// Transformers defines a list of transformer applied in order
type Transformers []Transformer

// Apply applies a single transformation based on a key
func (t *Transformer) Apply(input string) (output string, err error) {

	if input == "" {
		return "", ErrEmptyInput
	}

	output = input

	if len(t.AddPrefix) > 0 {
		output = fmt.Sprintf("%s%s", t.AddPrefix, output)
	}

	if len(t.AddSuffix) > 0 {
		output = fmt.Sprintf("%s%s", output, t.AddSuffix)
	}

	if len(t.TrimPrefix) > 0 {
		output = strings.TrimPrefix(output, t.TrimPrefix)
	}

	if len(t.TrimSuffix) > 0 {
		output = strings.TrimSuffix(output, t.TrimSuffix)
	}

	if len(t.Replacers) > 0 {
		r := strings.NewReplacer(t.Replacers.Unmarshal()...)
		output = r.Replace(output)
	}

	if t.Replacer != (Replacer{}) {
		r := strings.NewReplacer(t.Replacer.Unmarshal()...)
		output = r.Replace(output)
	}

	if len(t.Find) > 0 {
		re, err := regexp.Compile(t.Find)
		if err != nil {
			return "", err
		}

		output = re.FindString(output)
	}

	if t.FindSubMatch != (FindSubMatch{}) {
		output, err = t.FindSubMatch.Apply(output)
		if err != nil {
			return "", err
		}
	}

	if len(t.SemVerInc) > 0 {
		output, err = applySemVerInc(output, t.SemVerInc)
		if err != nil {
			return "", err
		}
	}

	if t.Quote {
		output = fmt.Sprintf("%q", output)
	}

	if t.Unquote {
		output = strings.Trim(output, "\"")
	}

	if t.JsonMatch != (JsonMatch{}) {
		var data any
		var results []any

		err = json.Unmarshal([]byte(output), &data)
		if err != nil {
			return "", err
		}

		queryResult, err := daselV2.Select(data, t.JsonMatch.Key)
		if err != nil {
			// if we had an error, and it's NOT a prop not found then we should fail
			if !strings.Contains(err.Error(), "could not access map index: property not found") {
				return "", err
			}
			// If the query result is not found, then we return an empty array
			results = []any{}
		} else {
			results = queryResult.Interfaces()
		}

		if len(results) == 0 {
			if t.JsonMatch.NoMatchResult == "<input>" {
				return output, nil
			} else if t.JsonMatch.NoMatchResult == "<blank>" {
				return "", nil
			} else if len(t.JsonMatch.NoMatchResult) > 0 {
				return t.JsonMatch.NoMatchResult, nil
			} else {
				err = fmt.Errorf("could not find value for query %q", t.JsonMatch.Key)
				return "", err
			}
		}
		if len(results) > 1 {
			if len(t.JsonMatch.JoinMultipleMatches) > 0 {
				stringResults := make([]string, len(results))
				for k, v := range results {
					stringResults[k] = fmt.Sprint(v)
				}
				return strings.Join(stringResults, t.JsonMatch.JoinMultipleMatches), nil
			} else if t.JsonMatch.MultipleMatchSelector == "first" {
				return fmt.Sprint(results[0]), nil
			} else if t.JsonMatch.MultipleMatchSelector == "last" {
				return fmt.Sprint(results[len(results)-1]), nil
			} else if len(t.JsonMatch.MultipleMatchSelector) > 0 {
				var index int
				_, err := fmt.Sscanf(t.JsonMatch.MultipleMatchSelector, "[%d]", &index)
				if err != nil {
					return "", err
				}
				if index < -(len(results)) || index >= len(results) {
					err = fmt.Errorf("selector out of range for query %q (%d vs %d)", t.JsonMatch.Key, index, len(results))
					return "", err
				}
				if index < 0 {
					index = len(results) + index
				}
				return fmt.Sprint(results[index]), nil
			} else {
				err = fmt.Errorf("multiple results found for query %q", t.JsonMatch.Key)
				return "", err
			}
		}
		output = fmt.Sprint(results[0])
		for _, v := range results {
			output = fmt.Sprint(v)
		}
	}

	return output, nil
}

// Apply applies a list of transformers
func (t *Transformers) Apply(input string) (string, error) {
	output := input

	err := t.Validate()
	if err != nil {
		return "", err
	}

	logrus.Info("[transformers]\n")

	for _, transformer := range *t {
		previous := output
		output, err = transformer.Apply(output)
		if err != nil {
			return "", err
		}

		logrus.Infof("✔ Result correctly transformed from %q to %q", previous, output)
	}
	return output, nil
}

func applySemVerInc(input, semVerInc string) (string, error) {

	if len(semVerInc) == 0 {
		return "", fmt.Errorf("no incremental semantic versioning rule, accept comma separated list of major,minor,patch")
	}

	v, err := semver.NewVersion(input)
	if err != nil {
		return "", fmt.Errorf("wrong semantic version input: %q", input)
	}

	rules := strings.Split(semVerInc, ",")
	for _, rule := range rules {
		switch rule {
		case "major":
			*v = v.IncMajor()
		case "minor":
			*v = v.IncMinor()
		case "patch":
			*v = v.IncPatch()
		default:
			return "", fmt.Errorf("unsupported incremental semantic versioning rule %q, only accept a comma separated list between major, minor, patch", semVerInc)
		}
	}
	return v.String(), nil

}

func (t *Transformer) Validate() error {

	warningMessageToLowerCase := func(key string) {
		logrus.Warningf("%q is deprecated in favor of %q", key, strings.ToLower(key))
	}

	warningMessageValueIgnore := func(key string) {
		logrus.Warningf("Key %q and %q are mutually exclusive, ignoring %q ", key, strings.ToLower(key), key)
	}

	if len(t.DeprecatedAddPrefix) > 0 {
		warningMessageToLowerCase("addPrefix")

		switch len(t.AddPrefix) {
		case 0:
			t.AddPrefix = t.DeprecatedAddPrefix
			t.DeprecatedAddPrefix = ""
		default:
			warningMessageValueIgnore("addPrefix")
		}

	}

	if len(t.DeprecatedAddSuffix) > 0 {
		warningMessageToLowerCase("addSuffix")
		switch len(t.AddSuffix) {
		case 0:
			t.AddSuffix = t.DeprecatedAddSuffix
			t.DeprecatedAddSuffix = ""
		default:
			warningMessageValueIgnore("addSuffix")
		}
	}

	if len(t.DeprecatedTrimPrefix) > 0 {
		warningMessageToLowerCase("trimPrefix")
		switch len(t.TrimPrefix) {
		case 0:
			t.TrimPrefix = t.DeprecatedTrimPrefix
			t.DeprecatedTrimPrefix = ""
		default:
			warningMessageValueIgnore("trimPrefix")
		}
	}

	if len(t.DeprecatedTrimSuffix) > 0 {
		warningMessageToLowerCase("trimSuffix")
		switch len(t.TrimSuffix) {
		case 0:
			t.TrimSuffix = t.DeprecatedTrimSuffix
			t.DeprecatedTrimSuffix = ""
		default:
			warningMessageValueIgnore("trimSuffix")
		}
	}

	if len(t.DeprecatedSemVerInc) > 0 {
		warningMessageToLowerCase("semverInc")
		switch len(t.SemVerInc) {
		case 0:
			t.SemVerInc = t.DeprecatedSemVerInc
			t.DeprecatedSemVerInc = ""
		default:
			warningMessageValueIgnore("semverInc")
		}
	}

	// t.DeprecatedFindSubMatch
	f := FindSubMatch{}
	value := t.DeprecatedFindSubMatch

	// If the manifest contains only the `pattern` string, then `0` is the implied value of `captureIndex`
	// Otherwise, both pattern and captureIndex are retrieved from the map value of the manifest
	// Note also that a value of `0` for `captureIndex` returns all submatches, and individual submatch indexes start at `1`.
	if _, ok := value.(string); ok {
		f.Pattern = value.(string)
		f.CaptureIndex = 0
	} else {
		err := mapstructure.Decode(value, &f)
		if err != nil {
			return err
		}

	}

	if f != (FindSubMatch{}) {
		warningMessageToLowerCase("findSubMatch")

		switch t.FindSubMatch == (FindSubMatch{}) {
		case true:
			t.FindSubMatch.Pattern = f.Pattern
			t.FindSubMatch.CaptureIndex = f.CaptureIndex
			t.DeprecatedFindSubMatch = nil
		case false:
			warningMessageValueIgnore("findSubMatch")
		default:
			logrus.Errorln("unexpected findsubmatch error")
		}
	}

	err := t.FindSubMatch.Validate()

	if err != nil {
		return err
	}

	return nil
}

func (t *Transformers) Validate() error {

	var errs []error

	transformers := *t

	for i, transformer := range transformers {
		err := transformer.Validate()
		if err != nil {
			errs = append(errs, err)
		}
		transformers[i] = transformer
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorln(e)
		}

		return errors.New("transformers validation failed")
	}
	return nil
}
