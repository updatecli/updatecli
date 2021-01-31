package source

// 2021/01/31
// Deprecated in favor of Transformer, need to be deleted in a future release

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
