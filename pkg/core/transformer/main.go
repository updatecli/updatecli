package transformer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

//Transformer holds a tranformer rule
type Transformer map[string]interface{}

//Transformers holds a list of transformer
type Transformers []Transformer

// Apply applies a single transformation based on a key
func (t *Transformer) Apply(input string) (output string, err error) {

	output = input

	for key, value := range *t {
		switch key {

		case "prefix":
			output = fmt.Sprintf("%s%s", value, input)

		case "suffix":
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
			logrus.Error(err)
			return "", err
		}
	}
	return output, nil
}
