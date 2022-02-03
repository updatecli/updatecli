package reports

// Stage contains the information if specific stage reported a success, a change, an error
type Stage struct {
	Name   string
	Kind   string
	Result string
}
