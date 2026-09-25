package transformer

// Replacer defines a text replacement.
type Replacer struct {
	// "from" defines the text to replace.
	//
	// example:
	//   * from: "_"
	//
	From string `yaml:",omitempty" jsonschema:"required"`
	// "to" defines the text replacing "from".
	//
	// example:
	//   * to: "."
	//
	To string `yaml:",omitempty" jsonschema:"required"`
}

// Replacers defines a list of text replacements.
type Replacers []Replacer

// Unmarshal read a struct of Replacers then return a slice of string
func (replacers Replacers) Unmarshal() (result []string) {

	for _, r := range replacers {
		result = append(result, r.From)
		result = append(result, r.To)
	}
	return result
}

// Unmarshal read a struct of Replacer then return a slice of string
func (r Replacer) Unmarshal() (result []string) {

	result = append(result, r.From)
	result = append(result, r.To)

	return result
}
