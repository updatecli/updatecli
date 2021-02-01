package transformer

// Replacer is struct used to feed strings.Replacer
type Replacer struct {
	From string
	To   string
}

// Replacers is an array of Replacer
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
