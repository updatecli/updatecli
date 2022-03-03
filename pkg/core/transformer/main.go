package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

//Transformer holds a tranformer rule
type Transformer map[string]interface{}

//Transformers holds a list of transformer
type Transformers []Transformer

// Apply applies a single transformation based on a key
func (t *Transformer) Apply(input string) (output string, err error) {

	if input == "" {
		return "", fmt.Errorf("Validation error: transformer input is empty.")
	}

	output = input

	for key, value := range *t {
		switch key {

		case "addPrefix":
			output = fmt.Sprintf("%s%s", value, input)

		case "addSuffix":
			output = fmt.Sprintf("%s%s", input, value)

		case "trimPrefix":
			val, ok := value.(string)

			if !ok {
				return "", fmt.Errorf("unknown value for trimPrefix: %v", val)
			}

			output = strings.TrimPrefix(input, val)

		case "trimSuffix":
			val, ok := value.(string)

			if !ok {
				return "", fmt.Errorf("unknown value for trimSuffix: %v", val)
			}

			output = strings.TrimSuffix(output, val)

		case "replacers":

			r := Replacers{}

			err := mapstructure.Decode(value, &r)
			if err != nil {
				return "", err
			}

			args := r.Unmarshal()

			replacer := strings.NewReplacer(args...)

			output = (replacer.Replace(output))
		case "replacer":

			r := Replacer{}

			err := mapstructure.Decode(value, &r)
			if err != nil {
				return "", err
			}

			args := r.Unmarshal()

			replacer := strings.NewReplacer(args...)

			output = (replacer.Replace(output))

		case "find":

			val, ok := value.(string)

			if !ok {
				return "", fmt.Errorf("unknown value for find: %v", val)
			}

			re, err := regexp.Compile(val)
			if err != nil {
				return "", err
			}

			found := re.FindString(output)

			output = found

		case "findSubMatch":

			f := FindSubMatch{}

			// Here I deal with 2 ways of providing the data within the yaml
			// only one parameter in string which is the pattern
			// considering 0 as the CaptureIndex default
			// or receiving both parameter as a map.
			if _, ok := value.(string); ok {
				f.Pattern = value.(string)
				f.CaptureIndex = 0
			} else {
				err := mapstructure.Decode(value, &f)
				if err != nil {
					fmt.Printf("Erreur de decodage %v\n", err)
					return "", err
				}
			}

			if len(f.Pattern) == 0 {
				return "", fmt.Errorf("no regex provided")
			}

			// Trying to compile the RegExp
			re, err := regexp.Compile(f.Pattern)
			if err != nil {
				return "", err
			}

			// Using the FindStringSubmatch function
			found := re.FindStringSubmatch(output)

			// I check if nothing found
			if len(found) == 0 {
				logrus.Debugf("No result found after applying regex %q to %q", f.Pattern, output)
				return "", nil
			}

			// I check if I got a capture with that index
			if len(found) <= f.CaptureIndex {
				logrus.Debugf("No capture found at position %v after applying regex %q to %q, full result with CaptureIndex 0 would be %v", f.CaptureIndex, f.Pattern, output, found)
				return "", nil
			}

			// here is the returned string
			output = found[f.CaptureIndex]

		case "semverInc":
			val, ok := value.(string)

			if !ok {
				return "", fmt.Errorf("unknown value for find: %v", val)
			}

			if len(val) == 0 {
				return "", fmt.Errorf("no incremental semantic versioning rule, accept comma separated list of major,minor,patch")
			}

			v, err := semver.NewVersion(input)
			if err != nil {
				return "", fmt.Errorf("wrong semantic version input: %q", val)
			}

			rules := strings.Split(val, ",")
			for _, rule := range rules {
				switch rule {
				case "major":
					*v = v.IncMajor()
				case "minor":
					*v = v.IncMinor()
				case "patch":
					*v = v.IncPatch()
				default:
					return "", fmt.Errorf("unsupported incremental semantic versioning rule %q, only accept a comma separated list between major, minor, patch", val)
				}
			}
			output = v.String()

		default:
			return "", fmt.Errorf("key '%v' not supported", key)
		}

	}

	return output, nil
}

// Apply applies a list of transformers
func (t *Transformers) Apply(input string) (output string, err error) {
	output = input
	for _, transformer := range *t {
		output, err = transformer.Apply(output)

		if err != nil {
			return "", err
		}
	}
	return output, nil
}
