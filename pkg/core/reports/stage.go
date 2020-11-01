package reports

// Stage contains the information if specific stage reported a success, a change, an error
type Stage struct {
	Name   string
	Kind   string
	Result string
}

// New init a new state object
func (s *Stage) New(kind, result string) {
	s.Kind = kind
	s.Result = result
}
